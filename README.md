<div align="center">
  <a href="https://github.com/scribe-org/Scribe-Server"><img src="https://raw.githubusercontent.com/scribe-org/Organization/main/logo/ScribeGitHubOrgBanner.png" width=1024 alt="Scribe Logo"></a>
</div>

[![platforms](https://img.shields.io/static/v1?message=Toolforge&logo=Wikidata&color=990000&logoColor=white&label=%20)](https://github.com/scribe-org/Scribe-Server)
[![issues](https://img.shields.io/github/issues/scribe-org/Scribe-Server?label=%20&logo=github)](https://github.com/scribe-org/Scribe-Server/issues)
[![language](https://img.shields.io/badge/Python%203-306998.svg?logo=python&logoColor=ffffff)](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md)
[![license](https://img.shields.io/github/license/scribe-org/Scribe-Server.svg?label=%20)](https://github.com/scribe-org/Scribe-Server/blob/main/LICENSE.txt)
[![coc](https://img.shields.io/badge/Contributor%20Covenant-ff69b4.svg)](https://github.com/scribe-org/Scribe-Server/blob/main/.github/CODE_OF_CONDUCT.md)
[![mastodon](https://img.shields.io/badge/Mastodon-6364FF.svg?logo=mastodon&logoColor=ffffff)](https://wikis.world/@scribe)
[![matrix](https://img.shields.io/badge/Matrix-000000.svg?logo=matrix&logoColor=ffffff)](https://matrix.to/#/#scribe_community:matrix.org)

## Toolforge service for Scribe app data downloads

**Scribe-Server** is a Toolforge service that provides the API by which data is available for download within Scribe apps. The goal is to create a [Scribe-Data](https://github.com/scribe-org/Scribe-Data) based regularly updating dataset that can signal new data availability as well as allow for language pack downloads.

> [!NOTE]\
> The [contributing](#contributing) section has information for those interested, with the articles and presentations in [featured by](#featured-by) also being good resources for learning more about Scribe.

Scribe apps are available on [iOS](https://github.com/scribe-org/Scribe-iOS), [Android](https://github.com/scribe-org/Scribe-Android) (planned) and [Desktop](https://github.com/scribe-org/Scribe-Desktop) (planned). For the data formatting processes see [Scribe-Data](https://github.com/scribe-org/Scribe-Data).

<a id="contents"></a>

# **Contents**

- [Contributing](#contributing)
- [Setup](#setup)
- [Supported Languages](#supported-languages)
- [Featured By](#featured-by)

<a id="contributing"></a>

# Contributing [`⇧`](#contents)

<a href="https://matrix.to/#/#scribe_community:matrix.org"><img src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/MatrixLogoGrey.png" height="50" alt="Public Matrix Chat" align="right"></a>

Scribe uses [Matrix](https://matrix.org/) for communications. You're more than welcome to [join us in our public chat rooms](https://matrix.to/#/#scribe_community:matrix.org) to share ideas, ask questions or just say hi :)

Please see the [contribution guidelines](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md) if you are interested in contributing to Scribe-Server. Work that is in progress or could be implemented is tracked in the [issues](https://github.com/scribe-org/Scribe-Server/issues) and [projects](https://github.com/scribe-org/Scribe-Server/projects).

> [!NOTE]\
> Just because an issue is assigned on GitHub doesn't mean that the team isn't interested in your contribution! Feel free to write [in the issues](https://github.com/scribe-org/Scribe-Server/issues) and we can potentially reassign it to you.

Those interested can further check the [`-next release-`](https://github.com/scribe-org/Scribe-Server/labels/-next%20release-) and [`-priority-`](https://github.com/scribe-org/Scribe-Server/labels/-priority-) labels in the [issues](https://github.com/scribe-org/Scribe-Server/issues) for those that are most important, as well as those marked [`good first issue`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) that are tailored for first time contributors.

### Ways to Help [`⇧`](#contents)

- [Reporting bugs](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=bug&template=bug_report.yml) as they're found 🐞
- Working on [new features](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Afeature) ✨
- [Localization](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Alocalization) for the app and Google Play 🌐
- [Documentation](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Adocumentation) for onboarding and project cohesion 📝
- Adding language data to [Scribe-Data](https://github.com/scribe-org/Scribe-Data/issues) via [Wikidata](https://www.wikidata.org/)! 🗃️

### Road Map [`⇧`](#contents)

The Scribe road map can be followed in the organization's [project board](https://github.com/orgs/scribe-org/projects/1) where we list the most important issues along with their priority, status and an indication of which sub projects they're included in (if applicable).

### Data Edits [`⇧`](#contents)

Scribe does not accept direct edits to the grammar JSON files as they are sourced from [Wikidata](https://www.wikidata.org/). Edits can be discussed and the [Scribe-Data](https://github.com/scribe-org/Scribe-Data) queries will be changed and ran before an update. If there is a problem with one of the files, then the fix should be made on [Wikidata](https://www.wikidata.org/) and not on Scribe. Feel free to let us know that edits have been made by [opening a data issue](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=data&template=data_wikidata.yml) or contacting us in the [issues for Scribe-Data](https://github.com/scribe-org/Scribe-Data/issues) and we'll be happy to integrate them!

<a id="setup"></a>

# Setup [`⇧`](#contents)

- Setup instructions forthcoming

<a id="supported-languages"></a>

# Supported Languages [`⇧`](#contents)

Scribe's goal is functional, feature-rich keyboards and interfaces for all languages. Check the [Scribe-Data extract_transform](https://github.com/scribe-org/Scribe-Data/tree/main/src/scribe_data/extract_transform) directory for queries for currently supported languages and those that have substantial data on [Wikidata](https://www.wikidata.org/).

The following table shows the supported languages and the amount of data available for each on [Wikidata](https://www.wikidata.org/) and via [Unicode CLDR](https://github.com/unicode-org/cldr) for emojis:

| Languages  |   Nouns | Verbs | Translations\* | Prepositions† | Emoji Keywords |
| :--------- | ------: | ----: | -------------: | ------------: | -------------: |
| French     |  17,072 | 6,572 |         67,652 |             - |          2,488 |
| German     | 102,833 | 3,593 |         67,652 |           210 |          2,898 |
| Italian    |   8,671 |    73 |         67,652 |             - |          2,457 |
| Portuguese |   5,437 |   536 |         67,652 |             - |          2,327 |
| Russian    | 194,448 |    12 |         67,652 |            15 |          3,827 |
| Spanish    |  39,105 | 4,930 |         67,652 |             - |          3,134 |
| Swedish    |  45,259 | 4,501 |         67,652 |             - |          2,913 |

`*` Given the current **`beta`** status where words are machine translated.

`†` Only for languages for which preposition annotation is needed.

<a id="featured-by"></a>

# Featured By [`⇧`](#contents)

<details open><summary><strong>Articles and Presentations on Scribe</strong></summary>
<p>

<strong>2023</strong>

- June: [Scribe iOS development blog post on Nested UITableViews & Apple's built-in ViewControllers in app menu](https://saurabhjamadagni.hashnode.dev/nested-uitableviews-apples-built-in-viewcontrollers) for [GSoC '23](https://www.mediawiki.org/wiki/Google_Summer_of_Code/2023#Accepted_projects:~:text=links%3A%20Phabricator%20issue-,3.%20Adding%20a%20Menu%20and%20Keyboards%20to%20Scribe%2DiOS,-%5Bedit%5D)
- March: [Presentation slides](https://docs.google.com/presentation/d/1W4ZkGi9UDDiTxM_silEij0gTE8YEubluHxe78xoqEP0/edit?usp=sharing) for a talk at [Berlin Hack and Tell](https://berlinhackandtell.rocks/) ([Hack of the month winner 🏆](https://berlinhackandtell.rocks/2023-03-28-no87-moore-hacks))

<strong>2022</strong>

- August: [Presentation slides](https://docs.google.com/presentation/d/12WNSt5xgNIAmSxPfvjno9-sBMGlvxG_xSaAxmHQDRNQ/edit?usp=sharing) for a session at the [2022 Wikimania Hackathon](https://wikimania.wikimedia.org/wiki/2022:Hackathon)
- July: [Presentation slides](https://docs.google.com/presentation/d/10Ai0-b8XUj5u9Hw4UgBtB7ufiPhvfFrb1vEUEyXYr5w/edit?usp=sharing) for a talk at [CocoaHeads Berlin](https://www.meetup.com/cocoaheads-berlin/)
- July: [Video on Scribe](https://www.youtube.com/watch?v=4GpFN0gGmy4&list=PL66MRMNlLyR7p9wsYVfuqJOjKZpbuwp8U&index=6) for [Wikimedia Celtic Knot 2022](https://meta.wikimedia.org/wiki/Celtic_Knot_Conference_2022)
- June: [Presentation slides](https://docs.google.com/presentation/d/1K2lj8PPgdx12I-xuhm--CBLrGm-Cz50NJmbp96zpGrk/edit?usp=sharing) for a talk with the [LD4 Wikidata Affinity Group](https://www.wikidata.org/wiki/Wikidata:WikiProject_LD4_Wikidata_Affinity_Group)
- June: [Scribe](https://github.com/scribe-org) featured for new developers on [MediaWiki](https://www.mediawiki.org/wiki/New_Developers)
- May: [Presentation slides](https://docs.google.com/presentation/d/1Cu3VwQ3lJUp5W84YDe0AFYS-6zfBxKsm0MI-OMl_IzY/edit?usp=sharing) for [Wikimedia Hackathon 2022](https://www.mediawiki.org/wiki/Wikimedia_Hackathon_2022)
- March: [Blog post](https://tech-news.wikimedia.de/en/2022/03/18/lexicographical-data-for-language-learners-the-wikidata-based-app-scribe/) on [Scribe-iOS](https://github.com/scribe-org/Scribe-iOS) for [Wikimedia Tech News](https://tech-news.wikimedia.de/en/homepage/) ([DE](https://tech-news.wikimedia.de/2022/03/18/sprachenlernen-mit-lexikografische-daten-die-wikidata-basierte-app-scribe/) / [Tweet](https://twitter.com/wikidata/status/1507335538596106257?s=20&t=YGRGamftI-5B_VwQ_bFRhA))
- March: [Presentation slides](https://docs.google.com/presentation/d/16ld_rCbwJCiAdRrfhF-Fq9Wm_ciHCbk_HCzGQs6TB1Q/edit?usp=sharing) for [Wikidata Data Reuse Days 2022](https://diff.wikimedia.org/event/wikidata-data-reuse-days-2022/)

</p>
</details>

<div align="center">
  <br>
    <a href="https://tech-news.wikimedia.de/en/2022/03/18/lexicographical-data-for-language-learners-the-wikidata-based-app-scribe/"><img height="120"src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikimediaDeutschlandLogo.png" alt="Wikimedia Deutschland logo linking to an article on Scribe in the tech news blog."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://www.mediawiki.org/wiki/New_Developers"><img height="120" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/MediawikiLogo.png" alt="MediaWiki logo linking to the new developers page."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://summerofcode.withgoogle.com/"><img height="120" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/GSoCLogo.png" alt="Google Summer of Code logo linking to its website."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <br>
</div>

# Powered By

### Contributors

Many thanks to all the [Scribe-Server contributors](https://github.com/scribe-org/Scribe-Server/graphs/contributors)! 🚀

<a href="https://github.com/scribe-org/Scribe-Server/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=scribe-org/Scribe-Server" />
</a>

### Wikimedia Communities

<div align="center">
  <br>
  <a href="https://www.wikidata.org/"><img height="175" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikidataLogo.png" alt="Wikidata logo"></a>
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <a href="https://www.wikipedia.org/"><img height="190" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikipediaLogo.png" alt="Wikipedia logo"></a>
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <br>
</div>
