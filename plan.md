# Plan d'implémentation pour l'exporteur Slurm

## Instructions pour l'IA de développement

**Contexte** : Nous allons développer un exporteur Prometheus pour Slurm en Go. Ce document contient l'architecture, la structure des dossiers et toutes les étapes d'implémentation.

**Rôle** : Agis comme un expert Go (Golang) respectant les standards de la "Clean Architecture".

**Directives importantes** :
- Suivre strictement la structure de projet définie dans ce document
- Implémenter une fonctionnalité complète à la fois
- Compiler et valider après chaque étape
- Faire un commit Git descriptif après chaque étape validée
- Écrire des commentaires clairs en anglais comme un humain le ferait
- Optimiser les requêtes pour minimiser les coûts

## Objectif
Créer un exporteur pour les métriques OpenMetrics natives de Slurm (version 25.11+).

## Architecture
L'exporteur sera composé des éléments suivants :
1. **Collecteur de métriques** : Interagit avec l'API de Slurm pour récupérer les métriques OpenMetrics.
2. **Serveur HTTP** : Expose les métriques au format Prometheus avec support pour Basic Auth et SSL.
3. **Gestion des erreurs** : Mécanismes pour gérer les erreurs de connexion ou de récupération des métriques, incluant des timeouts et des métriques d'erreur.
4. **Configuration** : Fichier de configuration pour définir les paramètres de connexion à Slurm, ainsi que les paramètres de sécurité (Basic Auth, SSL), les timeouts, et les labels personnalisés globaux.
5. **Logging** : Utilisation de `slog/log` pour le logging.
6. **Parsing des arguments** : Utilisation de `kingpin v2` pour le parsing des arguments.

## Endpoints et métriques Slurm
Slurm expose plusieurs endpoints pour récupérer les métriques OpenMetrics :
- `/metrics/jobs` : Métriques sur les jobs (états, CPU, mémoire, etc.).
- `/metrics/nodes` : Métriques sur les nœuds (CPU, mémoire, états, etc.).
- `/metrics/partitions` : Métriques sur les partitions (jobs, nœuds, états, etc.).
- `/metrics/jobs-users-accts` : Métriques sur les jobs par utilisateur et compte.
- `/metrics/scheduler` : Métriques sur le planificateur (cycles, threads, etc.).

Chaque endpoint retourne des métriques au format OpenMetrics, avec des labels pour identifier les ressources (par exemple, `node`, `partition`, `username`, `account`).

**Note** : Des exemples de outputs OpenMetrics natifs de Slurm sont disponibles dans le dossier `test_data` pour faciliter les tests et la validation. Les fichiers suivants sont disponibles :
- `test_data/metrics_jobs.txt`
- `test_data/metrics_nodes.txt`
- `test_data/metrics_partitions.txt`
- `test_data/metrics_jobs_users_accts.txt`
- `test_data/metrics_scheduler.txt`

## Étapes d'implémentation

### 1. Créer la structure du projet
- Initialiser un nouveau projet Go avec `go mod init`
- Créer une structure de répertoire appropriée : `cmd/`, `internal/`, `pkg/`, `test/`, etc.
- Ajouter un fichier `README.md` pour documenter le projet
- Créer un fichier `go.mod` et `go.sum` pour la gestion des dépendances

### 2. Configurer l'environnement et les dépendances
- Créer un fichier de configuration YAML pour définir les paramètres
- Ajouter les dépendances Go nécessaires :
  - `github.com/prometheus/client_golang` pour les métriques Prometheus
  - `gopkg.in/yaml.v3` pour parser la configuration YAML
  - `github.com/alecthomas/kingpin/v2` pour le parsing des arguments
  - Packages de la stdlib : `log/slog`, `net/http`, `crypto/tls`, etc.
- Créer les structures de configuration en Go correspondant au YAML
- Implémenter le chargement et la validation de la configuration
- Définir une variable globale Version (initialisée à dev) qui sera injectée via -ldflags lors de la compilation.

### 3. Implémenter le collecteur de métriques
- Développer un module pour interagir avec l'API de Slurm et récupérer les métriques OpenMetrics
- Implémenter un client HTTP avec support pour les timeouts et retry logic
- Parser les métriques OpenMetrics reçues de Slurm et les convertir en métriques Prometheus
- Ajouter les labels personnalisés globaux à toutes les métriques
- Interroger les endpoints suivants pour récupérer les métriques :
  - `/metrics/jobs` : Métriques sur les jobs
  - `/metrics/nodes` : Métriques sur les nœuds
  - `/metrics/partitions` : Métriques sur les partitions
  - `/metrics/jobs-users-accts` : Métriques sur les jobs par utilisateur et compte
  - `/metrics/scheduler` : Métriques sur le planificateur
- Implémenter un cache optionnel pour éviter de surcharger l'API Slurm

### 4. Créer le serveur HTTP
- Développer un serveur HTTP léger pour exposer les métriques au format Prometheus.
- Configurer les endpoints nécessaires
  - /metrics : Expose les métriques agrégées.
  - / : Landing Page HTML simple pointant vers /metrics et affichant la version.
- Implémenter le support pour Basic Auth et SSL pour sécuriser l'accès aux métriques.

### 5. Gestion des erreurs et logging
- Ajouter des mécanismes pour gérer les erreurs de connexion ou de récupération des métriques.
- Implémenter des logs avec `slog/log` pour surveiller le fonctionnement de l'exporteur.
- Ajouter des timeouts pour les requêtes vers l'API Slurm.
- Exposer des métriques d'erreur pour surveiller les échecs de récupération des métriques.
- Exposer la métrique `slurm_exporter_build_info` contenant la version du binaire.

### 6. Parsing des arguments
- Utiliser `kingpin v2` pour parser les arguments de la ligne de commande.
- Configurer les options pour le serveur HTTP, les paramètres de sécurité, et les logs.

### 7. Tests et validation
- Écrire des tests unitaires pour valider le fonctionnement du collecteur et du serveur.
- Tester l'intégration avec un environnement Slurm réel ou simulé.

### 8. Documentation
- Documenter l'installation, la configuration et l'utilisation de l'exporteur.
- Ajouter des exemples de configuration et d'utilisation.

### 9. Création d'un Makefile
- Créer un fichier `Makefile` avec les cibles suivantes :
  - `build` : Compiler l'exporteur en injectant la version et le commit sha via -ldflags.
  - `test` : Exécuter les tests unitaires
  - `clean` : Nettoyer les fichiers temporaires
  - `lint` : Linter le code (optionnel mais recommandé)
  - `run` : Lancer l'exporteur avec une configuration par défaut

### 10. Création d'une GitHub Action
- Configurer une GitHub Action dans `.github/workflows/` pour :
  - Builder et tester le code sur chaque push/PR
  - Créer des releases automatiques lors de la création de tags
  - Builder des binaires pour différentes plateformes (Linux, Windows, macOS)
  - Publier des releases sur GitHub avec les binaires attachés


## Structure de projet recommandée

```
slurm_exporter/
├── cmd/
│   └── slurm_exporter/
│       └── main.go              # Point d'entrée principal
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration et parsing YAML
│   ├── collector/
│   │   └── slurm.go           # Collecteur de métriques Slurm
│   ├── server/
│   │   └── http.go            # Serveur HTTP avec Basic Auth/SSL
│   └── metrics/
│       └── registry.go        # Registry des métriques Prometheus
├── pkg/                        # Packages publics (si nécessaire)
├── test_data/                  # Données de test Slurm
│   ├── metrics_jobs.txt
│   ├── metrics_nodes.txt
│   └── ...
├── configs/                    # Exemples de configuration
│   └── config.yaml
├── .github/
│   └── workflows/
│       └── release.yml         # GitHub Actions
├── Makefile                    # Commandes de build
├── go.mod                      # Dépendances Go
├── go.sum
├── README.md
└── LICENSE
```

## Exemple de fichier de configuration
Voici un exemple de fichier de configuration pour l'exporteur Slurm :

```yaml
# Configuration pour la connexion à l'API Slurm
slurm:
  url: "http://localhost:6817"
  timeout: "10s"  # Timeout pour les requêtes vers l'API Slurm

# Configuration du serveur HTTP
server:
  port: 8080
  basic_auth:
    enabled: true
    username: "admin"
    password: "password"
  ssl:
    enabled: false
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"

# Configuration des endpoints à exposer
endpoints:
  - name: "jobs"
    path: "/metrics/jobs"
    enabled: true
  - name: "nodes"
    path: "/metrics/nodes"
    enabled: true
  - name: "partitions"
    path: "/metrics/partitions"
    enabled: true
  - name: "jobs-users-accts"
    path: "/metrics/jobs-users-accts"
    enabled: true
  - name: "scheduler"
    path: "/metrics/scheduler"
    enabled: true

# Configuration des labels personnalisés globaux
labels:
  cluster: "cluster01"
  env: "prod"
  region: "eu-west-1"

# Configuration du logging
logging:
  level: "info"
  output: "stdout"
```

## Documentation et bonnes pratiques

### README.md
Créer un fichier `README.md` avec les informations suivantes :
- Une description claire et concise du projet.
- Des instructions d'installation et de configuration.
- Des exemples d'utilisation et de configuration.
- Des emojis pour rendre le document plus agréable à lire.
- Des informations sur les dépendances et les prérequis.

Exemple de structure pour le `README.md` :

```markdown
# Slurm Exporter 🚀

A Prometheus exporter for Slurm metrics, because monitoring your HPC cluster should be as smooth as your jobs running on it! 🚀

## Features
- ✅ Export Slurm metrics in OpenMetrics format
- ✅ Support for Basic Auth and SSL
- ✅ Customizable labels for all metrics
- ✅ Easy configuration with YAML

## Installation

### Prerequisites
- Go 1.20+
- Slurm 25.11+

### Build
```bash
git clone https://github.com/yourusername/slurm_exporter.git
cd slurm_exporter
make build
```

## Configuration

Create a `config.yaml` file with your settings:

```yaml
slurm:
  url: "http://localhost:6817"
  timeout: "10s"

server:
  port: 8080
  basic_auth:
    enabled: true
    username: "admin"
    password: "password"

labels:
  cluster: "cluster01"
  env: "prod"
```

## Usage

Run the exporter:

```bash
./slurm_exporter --config config.yaml
```

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License

MIT
```

### Commentaires dans le code
Les commentaires dans le code doivent être clairs, concis et utiles. Voici quelques bonnes pratiques :

1. **Commentaires de fonction** : Expliquer le but de la fonction, ses paramètres et son retour.
   ```go
   // fetchMetrics retrieves metrics from the Slurm API for a given endpoint.
   // It takes an endpoint as input and returns the metrics in OpenMetrics format.
   // If an error occurs during the request, it returns an error.
   func fetchMetrics(endpoint string) (string, error) {
       // Function implementation
   }
   ```

2. **Commentaires de logique complexe** : Expliquer les parties complexes du code.
   ```go
   // We use a mutex here to ensure thread-safe access to the metrics cache.
   // This prevents race conditions when multiple goroutines try to update the cache simultaneously.
   var cacheMutex sync.Mutex
   ```

3. **Commentaires de configuration** : Expliquer les options de configuration.
   ```go
   // BasicAuth configuration for securing the metrics endpoint.
   // If enabled, clients must provide a valid username and password to access the metrics.
   BasicAuth:
     Enabled: true
     Username: "admin"
     Password: "password"
   ```

## Bonnes pratiques de développement

### Optimisation des requêtes IA
Pour minimiser les coûts lors du développement avec l'IA, suivre ces bonnes pratiques :

1. **Travail par étapes courtes** : Implémenter et tester une fonctionnalité à la fois
2. **Messages concis et précis** : Donner des instructions claires et spécifiques
3. **Utiliser les informations existantes** : Référencer le plan existant et les exemples fournis
4. **Valider rapidement** : Compiler et tester fréquemment pour détecter les erreurs tôt
5. **Éviter les retours en arrière** : Bien planifier avant d'implémenter
6. **Utiliser les données de test** : Tester avec les fichiers fournis dans `test_data/`

### Commits Git
À chaque étape du développement, un commit Git devra être effectué pour suivre les modifications et faciliter la collaboration. Les messages de commit doivent être clairs et descriptifs, sans utiliser de caractères spéciaux comme ` qui pourraient être mal interprétés par GitHub.

Exemples de messages de commit :
- `feat: add basic auth support`
- `fix: handle timeout errors in metrics collection`
- `docs: update README with installation instructions`
- `test: add unit tests for metrics collector`

### Validation à chaque étape
Pour valider le fonctionnement à chaque étape, il est recommandé de compiler le code et de vérifier qu'il fonctionne comme attendu. Cela permet de détecter les erreurs tôt et de s'assurer que le code est toujours dans un état fonctionnel.

Exemple de commande pour compiler et exécuter les tests :
```bash
make build
make test
```

### Étapes de développement
1. **Implémenter la structure du projet** : Créer les répertoires et fichiers nécessaires.
   - Commit : `feat: initial project structure`
   - Validation : Compiler et vérifier que la structure est correcte.

2. **Configurer l'environnement et les dépendances** : Ajouter les dépendances et configurer l'environnement.
   - Commit : `feat: add dependencies and configure environment`
   - Validation : Compiler et vérifier que les dépendances sont correctement installées.

3. **Développer le collecteur de métriques** : Implémenter la logique pour récupérer les métriques.
   - Commit : `feat: implement metrics collector`
   - Validation : Compiler et tester avec les fichiers de test dans `test_data`.

4. **Créer le serveur HTTP** : Développer le serveur pour exposer les métriques.
   - Commit : `feat: implement HTTP server for metrics`
   - Validation : Compiler et vérifier que le serveur démarre correctement.

5. **Ajouter des métriques d'erreur et des timeouts** : Améliorer la robustesse de l'exporteur.
   - Commit : `feat: add error metrics and timeouts`
   - Validation : Compiler et tester les scénarios d'erreur.

6. **Configurer des labels personnalisés globaux** : Ajouter des métadonnées aux métriques.
   - Commit : `feat: add global custom labels`
   - Validation : Compiler et vérifier que les labels sont correctement ajoutés.

7. **Documenter le projet** : Ajouter un README et des commentaires dans le code.
   - Commit : `docs: add README and code comments`
   - Validation : Vérifier que la documentation est claire et complète.

## Prochaines étapes
- Implémenter la structure du projet.
- Configurer l'environnement et les dépendances.
- Développer le collecteur de métriques et le serveur HTTP.
- Les fichiers de test ont été créés dans le dossier `test_data` et peuvent être utilisés pour valider le fonctionnement de l'exporteur.
- Ajouter des métriques d'erreur et des timeouts pour améliorer la robustesse de l'exporteur.
- Configurer des modules pour chaque endpoint avec des intervalles de scrape personnalisés.
