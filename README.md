# Project Dump and Reconstruction

This project consists of a set of Go programs that enable the dumping of a project's directory structure and text file contents into a JSON file, reconstructing the project from the JSON file, and converting the JSON file into a Markdown representation of the project.
## Purpose

The purpose of this project is to provide a convenient way to package a project's source code and directory structure into a JSON format that can be easily shared, stored, or processed. This can be particularly useful for sharing source code with Large Language Models (LLMs) for analysis or assistance. The Markdown representation makes the source code and structure human-readable and suitable for documentation purposes.
## Files
- `common.go`: Contains common definitions and utility functions used by the other programs.
- `dump.go`: Dumps the project directory structure and text file contents into a JSON file.
- `reconstruct.go`: Reconstructs the project directory structure and file contents from a JSON file.
- `json2md.go`: Converts the JSON file into a Markdown representation of the project.
## Usage
### 1. Dumping the Project to JSON

The `dump.go` program scans the specified project directory, collects the directory structure and text file contents, and writes this information to a JSON file. It supports optional whitelist and blacklist patterns specified in `.gitignore`-formatted files.
#### Command

```sh
./dump -path=../your_project_directory -output=project_data.json -whitelist=whitelist.txt -blacklist=blacklist.txt
```


#### Flags
- `-path`: The root path of the project to be dumped. Default is the current directory.
- `-output`: The output JSON file. Default is `project_data.json`.
- `-whitelist`: Path to the whitelist file (optional). Only files matching these patterns will be included.
- `-blacklist`: Path to the blacklist file (optional). Files matching these patterns will be excluded.
### 2. Reconstructing the Project from JSON

The `reconstruct.go` program reads a JSON file containing a project’s directory structure and file contents, and reconstructs the project in the specified directory.
#### Command

```sh
./reconstruct -json=project_data.json -path=../reconstructed_project_directory
```


#### Flags
- `-json`: The input JSON file. Default is `project_data.json`.
- `-path`: The root path where the project will be reconstructed. Default is the current directory.
### 3. Generating Markdown from JSON

The `json2md.go` program reads a JSON file containing a project’s directory structure and file contents, and generates a Markdown file representing the project. The Markdown file includes the directory tree and the contents of text files formatted in code blocks.
#### Command

```sh
./json2md -json=project_data.json -output=project_structure.md
```


#### Flags
- `-json`: The input JSON file. Default is `project_data.json`.
- `-output`: The output Markdown file. Default is `project_structure.md`.
## Compilation

To compile the programs, run the following commands:

```sh
go build -o dump dump.go common.go
go build -o reconstruct reconstruct.go common.go
go build -o json2md json2md.go common.go
```


## Example Workflow
1. **Dump the project to JSON:**

```sh
./dump -path=../your_project_directory -output=project_data.json -whitelist=whitelist.txt -blacklist=blacklist.txt
``` 
2. **Reconstruct the project from JSON:**

```sh
./reconstruct -json=project_data.json -path=../reconstructed_project_directory
``` 
3. **Generate Markdown from the JSON file:**

```sh
./json2md -json=project_data.json -output=project_structure.md
```
## Project Structure

```go
.
├── common.go
├── dump.go
├── reconstruct.go
└── json2md.go
```


## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue to discuss any changes or improvements.
## License

This project is licensed under the MIT License.---

This `README.md` file provides detailed information about the purpose, usage, and structure of your project, along with instructions on how to compile and run the programs.