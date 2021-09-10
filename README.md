# Crosh: minimal CROss-platform SHell (WIP, code is not real yet)

Crosh is a minimal Bash-like language for cross-platform scripting on
Windows, macOS, and Linux. The goal is to cover up some of the major
differences between PowerShell and Bash like behavior of mv, cp, basic
file manipulation, environment variables, and basic control flow.

Fundementally though, it must be similar enough to Bash that it's not
easier to go learn a foreign cross-platform shell like
[Xonsh](https://xon.sh/),
[Nushell](https://github.com/nushell/nushell), or
[scsh](https://scsh.net/). Though all of these are more mature and
good choices, they require more investment to learn them since they
are less Bash-like.

## Non-Goals

* Support all of the Bash spec
* Be completely identical when cloning Linux userland programs like mv, cp
* Job control
* Pipes
* Complex expressions (array manipulation, string manipulation, arithmetic, etc.)

## WARNING 

Since it is new, it should be considered unsafe for anything but
containerized build environments where it doesn't matter so much if
the language implementation has bugs.

### Declarations

#### x = 1

#### export x = 1

This adds `X=1` to the environment variables passed to any execututions called.

### Subshells

#### $(ls ./file)

#### ?$(ls ./file)

### Control Flow

#### if $(eq "a" $var) ... elif $?--minify ... else ... endif

#### for file in $(ls ./dir) ... endfor

### Builtin Variables

#### $N

Arguments can be addressed like in Bash: `$0`, `$1`... `$N`.

#### $@

All arguments excluding `$0`.

#### $--my-flag, $-m, $-my-flag

Returns the command line argument following the flag.

#### $?--my-flag, $?-m, $?-my-flag

Returns true if the flag exists in command line arguments.

#### $MY_ENV

Returns the environment variable value or empty string if it is not set.

#### $?MY_ENV

Returns true if the environment variable has been set.

### Builtin Functions

#### exit $code

#### which gcc

#### prepend "string\n" ./file

This modifies the file, adding the prefix string.

#### append "\nstring\n" ./file

This modifies the file, adding the suffix string.

#### replace "string-or-regexp" ./file

#### mv ./from-file ./to-file

#### rm ./file ./or-directory

#### cp ./file ./or-directory

#### cd' ./directory

#### eq $a $b

#### neq $a $b
