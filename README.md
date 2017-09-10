# robospider
[![license](https://img.shields.io/badge/license-MIT-blue.svg)]()

A command line spider for robot.txt files written in go, highly inspired from [Parsero](https://tools.kali.org/information-gathering/parsero).

## Installation
You can install using go __get__ command, this will download and build the project for you:
```
go get github.com/b4dnewz/robospider
```
Now, if you have `$GOPATH/bin` in your __PATH__ you can start using it by typing __robospider__.

---

## How does it work?
Once is called with a target domain it will read the `robots.txt` file of a web server and looks at the Disallow entries. The Disallow entries tell the search engines what directories or files hosted on a web server mustn't be indexed. This is the way the administrator have to not share sensitive or private information with the search engines.

The steps involved in the scan are:
1. It will try to locate and read the `robots.txt` file.
2. For every Disallow entry found it will check if is reachable.
3. All the interesting results will be grouped into an output file.

In next versions it will perform a search for found addresses on web search engines and gather results.

Also, it doesn't mean that the files or directories typed in the Dissallow entries will not be indexed by Bing, Google, Yahoo, etc. For this reason, __robospider__ is capable of searching in Bing to locate content indexed without the web administrator authorization. So in next versions __robospider__ will perform additional check using search engines results.

---

## Usage
The last argument __must__ be the domain or the url to scan.
```

   |  |
   \**/    Robospider v1.0.0
  o={}=o   by Filippo 'b4dnewz' Conti
 / /()\ \  codekraft-studio <info@codekraft.it>
   \  /
 
Package usage: robospider [-proxy URL] [-output NAME] [DOMAIN]
 
  -output string
        the output file name Default: [domain].log
  -proxy string
        the full address of the proxy server to use: [address:port]

```
---

## Contributing

1. Create an issue and describe your idea
2. Fork the project (https://github.com/b4dnewz/robospider/fork)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Publish the branch (`git push origin my-new-feature`)
6. Create a new Pull Request
