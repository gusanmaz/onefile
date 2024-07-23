# OneFile: Comprehensive Project and Package Management Tool

OneFile is a versatile command-line tool designed to streamline various operations on project structures and package contents. It enables developers to easily dump local projects, reconstruct projects from JSON, convert JSON to Markdown, and fetch contents from GitHub repositories and PyPI packages.

## Features

- **Local Project Dumping**: Convert local project structures and contents to JSON or Markdown.
- **Project Reconstruction**: Rebuild project structures from JSON files.
- **JSON to Markdown Conversion**: Transform JSON project representations into readable Markdown format.
- **GitHub Repository Fetching**: Retrieve and save GitHub repository structures and contents.
- **PyPI Package Fetching**: Download and save PyPI package structures and contents.
- **Flexible Output**: Choose between JSON and Markdown output formats.
- **Progress Reporting**: View download progress for fetching operations.
- **Customizable Inclusion/Exclusion**: Use patterns to include or exclude specific files.
- **Git Integration**: Option to use git clone for faster repository fetching.
- **GitHub Token Support**: Authenticate with GitHub API for higher rate limits.

## Installation

### Option 1: Install as a single command

Ensure you have Go installed on your system, then run:

```bash
go install github.com/gusanmaz/onefile@latest
```

This installs the `onefile` binary in your `$GOPATH/bin` directory. Ensure this directory is in your system's PATH to run OneFile from anywhere.

### Option 2: Build from source

To build the project from source, clone the repository and use the provided build script:

```bash
git clone https://github.com/gusanmaz/onefile.git
cd onefile
chmod +x build.sh
./build.sh
```

This creates the main `onefile` executable and individual command executables in the `bin` directory.

## Usage

OneFile can be used either as a single command with subcommands or as individual commands.

### Single Command Usage

#### 1. Dumping a Local Project

```sh
onefile dump -p /path/to/your/project -o output_file -t json
```

Flags:
- `-p, --path`: Project root path (default: current directory)
- `-o, --output`: Output file name (without extension)
- `-t, --type`: Output type: 'json' or 'md' (default: 'json')
- `-e, --exclude`: Patterns to exclude files (space-separated)
- `--include-git`: Include .git files and directories
- `--include-non-text`: Include non-text files

#### 2. Reconstructing a Project from JSON

```sh
onefile reconstruct -j project_data.json -o /path/to/output/directory
```

Flags:
- `-j, --json`: Input JSON file
- `-o, --output`: Output directory for project reconstruction

#### 3. Converting JSON to Markdown

```sh
onefile json2md -j project_data.json -o project_structure.md
```

Flags:
- `-j, --json`: Input JSON file
- `-o, --output`: Output Markdown file
- `--include-git`: Include .git files and directories
- `--include-non-text`: Include non-text files

#### 4. Fetching GitHub Repository

```sh
onefile github2file -u https://github.com/username/repo -t json -o output_file
```

Flags:
- `-u, --url`: Full GitHub repository URL
- `-t, --type`: Output type: 'json' or 'md' (default: 'md')
- `-o, --output-name`: Output file name (without extension)
- `-d, --output-dir`: Output directory
- `-e, --exclude`: Patterns to exclude files (space-separated)
- `-a, --all-repos`: Fetch all repositories for a user
- `-g, --use-git`: Use git clone if available (default: true)
- `-k, --token`: GitHub API token
- `--include-git`: Include .git files and directories
- `--include-non-text`: Include non-text files

#### 5. Fetching PyPI Package

```sh
onefile pypi2file -p package_name -t json -o output_file
```

Flags:
- `-p, --package`: PyPI package name
- `-t, --type`: Output type: 'json' or 'md' (default: 'md')
- `-o, --output-name`: Output file name (without extension)
- `-d, --output-dir`: Output directory
- `--include-git`: Include .git files and directories
- `--include-non-text`: Include non-text files

## Use Cases

1. **LLM Code Analysis**: Package entire projects for submission to Large Language Models for code review, refactoring suggestions, or documentation generation.
2. **Project Snapshots**: Create snapshots of project states for version control or backup purposes.
3. **Open Source Exploration**: Easily fetch and examine the structure of open-source projects on GitHub without cloning entire repositories.
4. **Documentation Generation**: Automatically generate project structure documentation in Markdown format for wikis or README files.
5. **Dependency Analysis**: Fetch PyPI packages to analyze their structure and contents before including them in your project.
6. **Code Sharing**: Share project structures and contents with colleagues or in forum posts without zipping and uploading entire projects.
7. **Project Comparisons**: Dump multiple project versions to JSON and use diff tools to compare structures and contents over time.
8. **Automated Tooling**: Incorporate OneFile into CI/CD pipelines for automated project analysis, documentation updates, or dependency checks.

## Contributing

Contributions to OneFile are welcome! Please feel free to submit pull requests, open issues, or suggest new features.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions about using OneFile, please open an issue on our GitHub repository.