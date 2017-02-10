# Welcome! [![Build Status](https://travis-ci.org/ccbrown/gggtracker.svg?branch=master)](https://travis-ci.org/ccbrown/gggtracker)

This is the repository for <a href="https://gggtracker.com" target="_blank">gggtracker.com</a>. If there's something you think the site is missing, please either a.) open an issue to request the feature or b.) develop the feature yourself and put in a pull request. Pull requests should be written in the same style as the existing code base and any new features should be implemented in a way that's as concise and unintrusive as possible. Before creating any pull requests, please run `make pre-commit` to format and test your changes.

### Running

If you have Go installed, you can run the server like so: `go get ./... && go run main.go`

Or if you just want to run a server using Docker: `docker run -it -p 8080:8080 ccbrown/gggtracker`

Either method will make a local instance of the tracker available at http://127.0.0.1:8080

### License / Attribution

This source is released under the MIT license (see the <i>LICENSE</i> file).

Some images used under Creative Commons 3.0 from this project: https://github.com/Templarian/WindowsIcons

This project is not affiliated with Grinding Gear Games.
