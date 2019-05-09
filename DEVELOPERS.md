# Developing Justitia

* [Development Setup](#setup)
* [Running Tests](#tests)
* [Coding Rules](#rules)
* [Commit Message Guidelines](#commits)

## <a name="setup"> Development Setup

This document describes how to set up your development environment to build and test Justitia, and
explains the basic mechanics of using `git`, `go` and `make`.

### Installing Dependencies

Before you can build Justitia, you must install and configure the following dependencies on your
machine:

* [Git](http://git-scm.com/): The [Github Guide to
  Installing Git][git-setup] is a good source of information.

* [Go v1.10.x](https://golang.org/doc/install): Go official installation documentation.

* [Make](https://git.savannah.gnu.org/git/make.git): Git clone source repository. Then follow the guide of `README.git`.

### Forking Justitia on Github

To contribute code to Justitia, you must have a GitHub account so you can push code to your own
fork of Justitia and open Pull Requests in the [GitHub Repository](https://github.com/DSiSc/justitia).

To create a Github account, follow the instructions [here](https://github.com/signup/free).
Afterwards, go ahead and [fork](http://help.github.com/forking) the
[main Justitia repository][github].


### Building Justitia

To build Justitia, you clone the source code repository and use `Make` to build the executable binary:

```shell
# Clone your Github repository:
git clone git@github.com:your_name_here/justitia.git

# Go to the Justitia directory:
cd justitia

# Fetch dependencies:
make fetch-deps

# Build Justitia:
make build
```

## Running Tests

### <a name="unit-tests"></a> Running the Unit Test Suite

To run tests, you should run:

```shell
make test
```

## <a name="rules"></a> Coding Rules

To ensure consistency throughout the source code, keep these rules in mind as you are working:

* All features or bug fixes **must be tested** by one or more unit-test.
* All public API methods **must be documented**
* All code **must be `gofmt`ed**
* Code coverage for pull request  **must exceed 60%**
* Authors **should avoid common mistakes** explained in the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) page, you can use `Golint` to help check code.

## <a name="rules"></a> Contribute Process

To contribute the code, you can follow the process bellow:

- New Feature
    - develop locally
      ```shell
      # create new branch:
      git checkout -b feature-xxx
      
      # develop new feature
      blabla
      
      # commit new feature
      git add xxx
      git commit -m 'commit comment'
      git push --set-upstream origin feature-xxx
      ```
    - create pull request to master branch on [Home Page](https://github.com/DSiSc/justitia)

- Bug Fix
    - develop locally
      ```shell
      # create new branch:
      git checkout -b bugfix-xxx 
      
      # develop new feature
      blabla
      
      # commit new feature
      git add xxx
      git commit -m 'commit comment'
      git push --set-upstream origin bugfix-xxx  
      ```
    - create pull request to master branch on [Home Page](https://github.com/DSiSc/justitia)

## <a name="commits"></a> Git Commit Guidelines

We have very precise rules over how our git commit messages can be formatted.  This leads to **more
readable messages** that are easy to follow when looking through the **project history**.

### Commit Message Format
Each commit message consists of a **header** and a **body**.  The header has a special
format that includes a **type** and a **subject**:

```
#header
<type># <subject>
<BLANK LINE>

#body
1. blabla
2. ...
```

Any line of the commit message cannot be longer 100 characters! This allows the message to be easier
to read on GitHub as well as in various git tools.

### Type
Must be one of the following:

* **feature**: A new feature
* **bugfix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing
  semi-colons, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **optimize**: A code change that improves performance
* **test**: Adding missing or correcting existing tests

### Subject
The subject contains succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize first letter
* no dot (.) at the end

### Body
Just as in the **subject**, use the imperative, present tense: "change" not "changed" nor "changes".
The body should include the motivation for the change and contrast this with previous behavior.