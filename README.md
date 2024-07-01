# Project Dump, Reconstruction, and GitHub Fetcher

This project consists of a set of Go programs that enable various operations on project structures and contents, including dumping a local project to JSON, reconstructing a project from JSON, converting JSON to Markdown, and fetching GitHub repositories.

## Purpose

The purpose of this project is to provide a convenient way to:
1. Package a local project's source code and directory structure into a JSON format.
2. Reconstruct a project from a JSON file.
3. Convert a project's JSON representation into a human-readable Markdown format.
4. Fetch a GitHub repository and save its structure and contents in either JSON or Markdown format.

These tools can be particularly useful for sharing source code with Large Language Models (LLMs) for analysis or assistance, creating project snapshots, or generating documentation.

## Files

- `common.go`: Contains common definitions and utility functions used by the other programs.
- `dump.go`: Dumps the project directory structure and text file contents into a JSON file.
- `reconstruct.go`: Reconstructs the project directory structure and file contents from a JSON file.
- `json2md.go`: Converts the JSON file into a Markdown representation of the project.
- `github2file.go`: Fetches a GitHub repository and creates either a JSON or Markdown file from its contents.

## Usage

### 1. Dumping a Local Project to JSON

The `dump.go` program scans the specified project directory, collects the directory structure and text file contents, and writes this information to a JSON file.

#### Command

```sh
./dump -path=../your_project_directory -output=project_data.json -whitelist=whitelist.txt -blacklist=blacklist.txt
```

#### Flags
- `-path`: The root path of the project to be dumped. Default is the current directory.
- `-output`: The output JSON file. Default is `project_data.json`.
- `-whitelist`: Path to the whitelist file (optional). Only files matching these patterns will be included.
- `-blacklist`: Path to the blacklist file (optional). Files matching these patterns will be excluded.

### 2. Reconstructing a Project from JSON

The `reconstruct.go` program reads a JSON file containing a project's directory structure and file contents, and reconstructs the project in the specified directory.

#### Command

```sh
./reconstruct -json=project_data.json -path=../reconstructed_project_directory
```

#### Flags
- `-json`: The input JSON file. Default is `project_data.json`.
- `-path`: The root path where the project will be reconstructed. Default is the current directory.

### 3. Generating Markdown from JSON

The `json2md.go` program reads a JSON file containing a project's directory structure and file contents, and generates a Markdown file representing the project.

#### Command

```sh
./json2md -json=project_data.json -output=project_structure.md
```

#### Flags
- `-json`: The input JSON file. Default is `project_data.json`.
- `-output`: The output Markdown file. Default is `project_structure.md`.

### 4. Fetching GitHub Repository and Creating JSON/Markdown

The `github2file.go` program fetches a GitHub repository and creates either a JSON or Markdown file containing the repository's structure and text file contents.

#### Command

```sh
./github2file -url=https://github.com/username/repo -type=json -output=output_file.json
```

or

```sh
./github2file -owner=username -repo=repository -type=md -output=output_file.md
```

#### Flags
- `-url`: The full GitHub repository URL (optional, can be used instead of -owner and -repo).
- `-owner`: The GitHub username of the repository owner (required if -url is not provided).
- `-repo`: The name of the GitHub repository (required if -url is not provided).
- `-type`: The output type, either 'json' or 'md' (default is 'json').
- `-output`: The name of the output file (optional, default is "{owner}_{repo}.{type}").

## Compilation

To compile the programs, run the following commands:

```sh
go build -o dump dump.go common.go
go build -o reconstruct reconstruct.go common.go
go build -o json2md json2md.go common.go
go build -o github2file github2file.go common.go
```

## Example Workflow

1. **Dump a local project to JSON:**

```sh
./dump -path=../your_project_directory -output=project_data.json
```

2. **Reconstruct the project from JSON:**

```sh
./reconstruct -json=project_data.json -path=../reconstructed_project_directory
```

3. **Generate Markdown from the JSON file:**

```sh
./json2md -json=project_data.json -output=project_structure.md
```

4. **Fetch a GitHub repository and create a JSON file:**

```sh
./github2file -url=https://github.com/octocat/Hello-World -type=json -output=github_project.json
```

5. **Fetch a GitHub repository and create a Markdown file:**

```sh
./github2file -owner=octocat -repo=Hello-World -type=md -output=github_project.md
```

## Features

- **Local Project Dumping**: Dump local project structure and contents to JSON.
- **Project Reconstruction**: Reconstruct project from JSON file.
- **JSON to Markdown Conversion**: Convert JSON project representation to readable Markdown format.
- **GitHub Repository Fetching**: Fetch GitHub repositories using full URL or owner/repo combination.
- **Progress Reporting**: Display download progress while fetching GitHub repositories.
- **Flexible Output**: Choose between JSON and Markdown output formats for GitHub fetching.
- **Error Handling**: Improved error handling for invalid GitHub URLs or non-existent repositories.

## Project Structure

```
.
├── common.go
├── dump.go
├── reconstruct.go
├── json2md.go
└── github2file.go
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue to discuss any changes or improvements.

## License

This project is licensed under the MIT License.