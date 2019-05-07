# Community
The heart of open source is the people who are participating in a community and it is no different in case of Justitia either.

## Communication.

Communication is underrated, but crucial when it comes to participating in an open source community. We have several tools that we use for online and offline communication, which we encourage you to use.

###Online Meeting
Every week, we have an online meeting where you can discuss not only project-related issues, but also any other issues. Meeting plan and address, we will update to justitia wiki page.

###Wiki Pages
Each project and working group has its own wiki page. Wiki pages are a secondary source of information. They hold information that is subject to more change than the info provided in the project’s documentation i.e. meeting agendas, team members, etc.

##Governance
The Justitia community has a few different roles for governance, leadership, and community participation. Each operates have different duty but being aware of each of them is useful.

###Roles

####Active User Contributor(AUC)
Individual Members who have contributed to Justitia project. AUCs is crucial for Justitia and their participation is highly encouraged.

####Active Technical Contributor(ATC)
Individual Members who have contributed to Justitia project over the last two 6-month release cycles are automatically considered as ATCs.

In specific cases you can apply for an exception to become an ATC, for further information please see the relevant section of the Technical Committee Charter.

####Active Project Contributor(APC)
If you have the ATC status, in the Justitia project where you contributed over the last two 6-month release cycles you are considered to be an Active Project Contributor.

Only APCs can participate in the election process to vote for the next PTL of the team.

####Project Team Lead (PTL)
Justitia have a Project Team Lead. She/He coordinate the day to day operation of the project, resolve technical disputes within the project, and operate as the spokesperson and ambassador for the project.
    
Project Team Leads are elected for each release cycle by Active Project Contributors: individuals who have contributed to the project in the last two release cycles.
    
####Core Reviewer
Justitia projects have a project team consisting of core reviewers and contributors.

Core reviewers are responsible for:

Defining and maintaining the project mission

Reviewing bug reports and deciding about their priority

Reviewing changes and approving them when it meets the design and coding or documentation standards of the project

Core reviewers have rights that blocking or approving a commit.

New core reviewers are elected by voting from the members of the core team of the project.

##Releases
Justitia release a version every three month. For each version, PTL decides which features and patches will be released. 

###Schedule and Planning
The three-month cycle is divided into three phases.

For the first phase: PTL decide which features and patches will be released and publish the plan to wiki page. 

For the second phase: Contributor develop the scheduled features and patches and commit the new code to repo.

For the third phase: One week before the release version, we will focus on testing and stop merging any commit except the bugfix.

###Tags
Once a 3-month development cycle is completed the code for that release is marked to a new release tag. When release a version, PTL creates new tag with the same annotate for each project(include sub-project) independently.

### Version Name
We use `X.Y.Z` as our version name format:
* X(Major version numbers) change whenever there is some significant change being introduced. For example, a large or potentially backward-incompatible change to a software package.
* Y(Minor version numbers) change when a new, minor feature is introduced or when a set of smaller features is rolled out.
* Z(Patch numbers) change when a new build of the software is released to customers. This is normally for small bug-fixes or the like.