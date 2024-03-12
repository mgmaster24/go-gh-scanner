# go-gh-scanner
GO application that scans GitHub for usage of desired information.

## Structure

### aws_sdk
This package is used to wrap AWS SDK related functionality.

#### Currently used SDK APIs:
- DynamoDB
  - Used to write results to a DynamoDB table provided in the application writer config
- SecretsManager
  - Used to retrieve secreted from AWS SecretsManager service

### cli
This package constructs a set of flags that can be passed to the application at runtime

#### Currently used flags
- -c
  - Used to provide the application configuration file location
  - Default value is `config.json`
  - Ex. go-gh-scanner -c path-to-config.json
- -t
  - Used to provide the location of the tokens json file
  - Default value is `tokens.json`
  - Ex. go-gh-scanner -t path-to-tokens.json

### config
This package defines the configuration objects and some helpful utilities for using the configuration
data elsewhere in the application

### github_api
This package abstracts the usage of the `github.com/google/go-github/github` package into easy to use methods
for scanning and retrieving code and repository information.

#### High Level Operations
- Code Scanning for Dependency usage
- Repository data retrieval
- Execution of query request of API url

### models
This package stored object definitions used through out the application

### reader
This package stores any "reading" functionality.  Could be from a file, database, etc.

#### Provided Functionality
- `ReadJsonData`
  - Generic funtion that will read JSON data from a file

### search
This packae is used to search the application for references to the provided tokens

#### Provided Functionality
- `FindTokenRefsInFiles`
  - Reads a files contents and search those contents for any references to the array of tokens provided.

### tokens
This package is used for retrieving tokens from an existing JSON token file.

#### Considerations
This functionality can be extended to retrieve tokens from any other location by conforming to the `TokenReader` interface.

### utils
This package contains a collection of utility methods providing additional helpful application functionality.

#### Current Functionality domains
- Archives
  - Extracting GZIP files
  - More to come
- HTTP
  - Create generic request objects
- OS
  - A handful of helpful file io methods
- Misc
  - A handful of helpful string search and/or manipulation methods

### writer
This package defines write operation interfaces to be extended for use by the specific areas of interest. (I.E. write to file, to database, etc.)


## Docker image
This application can be run as a container image.  The `DockerFile` can be altered to fit your own use case, but in a nutshell it will:
- Create an image using the `golang 1.22` docker image
- Copy all source and dependencied from this application into the image
- Copy the configuration file defined locally in your source tree.
  - NOTE: The configuration files are not provided when you clone/fork this repository.  They have to be created by the consumer of the application
  - If the file names differ from what is provided in the dockerfile then the dockerfile will need to be update to use your config file names
- Build the application
- Run the application providing the copied config file as arguments

## Applciation Configuration
The application currently uses a JSON definition for defining application functionality.

A list of application configuration variables are defined below:
- authTokenKey
  - This is the key used for retrieving the auth token from secrets manager
- owner
  - The owner of the organization or repos that the application wil scan for dependency usage
- extractDir
  - The directory to temporarily store repository archives
- languages
  - A list of languages and their file extension to search through
    - Language definitions consist of:
      - name (Ex. TypeScript, JavaScript, etc.)
      - extension (Ex. .ts, .js, etc.)
- packageFile
  - The file to scan for dependency info
- dependencies
  - The dependencied to search for in the package file defined above.
- reposToIgnore
  - A list of repos to not include in search results
- teamsToIgnore
  - A list of teams to not include in the `teams` result string
- tokenResultsWriterConfig
  - The configuration of the token results writer
    - Configuration
      - destination (table, file, etc.)
      - destinationType (table, file, etc.)
      - useBatch - use batch processing
- repoResultsWriterConfig
  - The configuration of the repo results writer
    - Configuration
      - destination (table, file, etc.)
      - destinationType (table, file, etc.)
      - useBatch - use batch processing

## Tokens
Tokens are used to search repository files for their use within the repository.  Tokens can be defined however you desire. This is accomplished by providing a token file or location of token data and defining a `TokenReader` for how to read the data on the tokens location.

## Running the Application
Follow the steps below to run the application
- Create an applciation config file.
  - See the [Application Config](#applciation-configuration) section above for more detail.
- Create a tokens reader
  - A user can define how they want to retrieve token data.  In its simplest form this can be a file with a list of strings.
  - Built in functionality reads tokens from a JSON file.
- Run `go run main.go -c path-to-app-config.json -t location-of-tokens`
- Or build `go build main.go`
  - Output is `go-gh-scanner`
  - Run `go-gh-scanner -c path-to-app-config.json -t location-of-tokens`

## Questions or Comments
If you have a question or comment, please feel free to open an issue in this repo and I will do my best to address it in a timely manner.