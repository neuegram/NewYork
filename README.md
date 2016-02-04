New York
========

### Configuration

There are several things that must be done to setup New York. 

First of all, you will need to edit **debug.bat** to replace `<command>` with the command you want to run. Feel free to move around `%1`, which is the filename argument passed from **nyc.go**. It is only placed where it is based on standard command line calling convention. An example of how **debug.bat** might look is as follows:

`windbg -G -Q -c "$$><events.wds;g" "program.exe" %1`

If the target process is in your path, feel free to shorten **debug.bat** to something like this:

`windbg -G -Q -c "$$><events.wds;g" program %1`

No changes are necessary to **nyc.go**, **cpu.bat**, or **events.wds**, but they are simple enough to alter if need be.

Secondly, make sure before running **nyc.exe** that **crash.log** does not exist in the same directory as the New York executable.

Third, you must have [!exploitable](https://msecdbg.codeplex.com/) installed properly.

Lastly, you will need a fileset. My scraper isn't up to par at the moment, but scraping [mmnt.ru](http://mmnt.ru/int) should suffice, especially when searching for obscure filetypes that aren't necessarily indexed by the major search engines. Your best bet is creating a minset by tracing codepaths with a tool such as [RunTracer](https://github.com/grugq/RunTracer/).

### Compilation

No external dependencies are required. Ensure you have the latest version of Go installed (version 1.5 at the time of writing). Versions 1.5+ have support for non-bootstrapped cross-compilation, so compiling on your host for use within a x86 or x64 VM should be no problem. To compile for your default architecture, run the following:

`go build neuebits.com/neuegram/newyork`

In order to cross-compile, you must set the environment variables **GOOS** and **GOARCH** to the corresponding values for the target. Then you may continue with running the command above.

### Usage

Generic: `nyc.exe <target process> <minset> <filetype> <timeout>`

Example: `nyc.exe program.exe minset .mp3 10`

### Process Termination

It's important to note that if `<timeout>` <= 0, then the process will be terminated when the process' CPU usage is 0. Otherwise, the process will be terminated after it has been detected as running (CPU usage not 0 or -1), in addition to the specified delay.

### TODO

+ Log errors to a file
+ Central server distribution of mutations
+ Central server controlled rolling restarts
+ Implement different methods for "dumb" fuzzing / mutation
+ Protocol fuzzing (in real time and from pcaps)
+ Linux support
+ Mac OS support
+ A lot more...

### Contact / Hire Me

Email Me: [neuegram@gmail.com](mailto:neuegram@gmail.com). Prefer PGP? [0x8ef4855f90493ec0](https://pgp.mit.edu/pks/lookup?op=get&search=0x8EF4855F90493EC0)