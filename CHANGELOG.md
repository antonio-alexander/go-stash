# Change Log

All notable changes to this service will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [1.0.3] - 04/09/23

- fixed bug with configuration that would cause panics
- updated workflow slightly

## [1.0.2] - 04/08/23

- updated the version in an attempt to satisfy go proxy since it was a private repository
- updated workflows

## [1.0.1] - 03/26/23

- fixed bug where initialized boolean wasn't unset on shutdown
- updated the logic in the memory stash such that it won't execute the eviction logic if the stash only has a single item in it
- updated go go 1.19
- added logger

## [1.0.0] - 09/27/22

- Initial version
