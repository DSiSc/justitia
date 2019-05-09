# Community
The heart of open source is the people who are participating in a community and it is no different in case of Justitia either.

## Communication.

Communication is underrated, but crucial when it comes to participating in an open source community. We have several tools that we use for online and offline communication, which we encourage you to use.

### Online Meeting
Every week, we have an online meeting where you can discuss not only project-related issues, but also any other issues. Meeting plan and address, we will update to justitia wiki page.

### Wiki Pages
Each project and working group has its own wiki page. Wiki pages are a secondary source of information. They hold information that is subject to more change than the info provided in the project’s documentation i.e. meeting agendas, team members, etc.

### Etherpad
We use `Etherpad` as our daily documentation tools, and you can share the documents, blogs, and articles with others in this `Pad`[Etherpad Address](https://etherpad.net/p/Justitia).

## Governance
The Justitia community has a few different roles for governance, leadership, and community participation. Each operates have different duty but being aware of each of them is useful.

### Roles

#### Individual Committer
Individual Members who have contributed to Justitia project.  Individual Committer is crucial for Justitia and their participation is highly encouraged.

#### Core Reviewer
Justitia projects have a project team consisting of core reviewers.

Core reviewers are responsible for:

Defining and maintaining the project mission

Reviewing bug reports and deciding about their priority

Reviewing changes and approving them when it meets the design and coding or documentation standards of the project

Core reviewers have rights that blocking or approving a commit.

#### Project Leader
Justitia have a Project Leader. She/He coordinate the day to day operation of the project, resolve technical disputes within the project, and operate as the spokesperson and ambassador for the project.

## Releases
Justitia release a version every three months. For each version, Project Leader decides which features and patches will be released. 

### Schedule and Planning
The three-month cycle is divided into three phases.

- For the first phase: 
    1. Project Leader decide which features and patches will be released and publish the plan to wiki page. 

- For the second phase: 
    1. Contributor develop the scheduled features and patches and commit the new code to repo.

- For the third phase:
    1. Two weeks before the release version, we will focus on testing and stop merging any commit except the bugfix.
    2. We will create a new tag for a batch of bugs, and these tags named in the format `{VersionName}.CR{Number}`(e.g. V1.0.0.CR1, V1.0.0.CR2...)
    3. If there is no high impact bugs, we will creates new tag with the annotate `{VersionName}` for each project(include sub-project) independently.
    
### Stable Branch
Once a three-month development cycle is completed, the code for that release is branched, in git, to a stable branch.

Stable branches are kept as a safe source of fixes for high impact bugs and security issues which have been fixed, on master, since the release occurred. Given the stable nature of these branches, backports to stable branches undergo additional scrutiny when they are proposed. Proposed changes should:

- Have a low risk of introducing a regression

- Have a user visible benefit

- Be self contained

- Be included in master and backported to all releases between master and the stable branch in question

Stable Branches are maintained until a new version was released.


### Version Name
We use `X.Y.Z` as our version name format:
* X(Major version numbers) change whenever there is some significant change being introduced. For example, a large or potentially backward-incompatible change to a software package.
* Y(Minor version numbers) change when a new, minor feature is introduced or when a set of smaller features is rolled out.
* Z(Patch numbers) change when a new build of the software is released to customers. This is normally for small bug-fixes or the like.
