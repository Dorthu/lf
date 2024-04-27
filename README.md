# lf

logfmt parser and filterer inspired by [`jq`](https://jqlang.github.io/jq/)

## Usage

Suppose your application outputs logs like this:

```
time="2024-04-26 20:56:00" msg="example message" tag=example
```

To get an easier-to-read format, you might pipe the log to `lf` like so:

```
tail -f /var/log/file | lf [.time] .msg
```

If you only wanted records with the tag "example2", you might do this:

```
tail -f /var/log/file | lf tag=example2
```

And to give output only matching lines in an easier format, you might say:

```
tail -f /var/log/file | lf 'tag=example | [.time] .msg'
```

`lf` provides two features: **Filtering** and **Formatting**

### Filtering

Filters are given as key/value pairs like normal logfmt, except that the `=` may be
one of four allowed operators:

| Operator | Meaning |
| --- | --- |
| `=` | Value must match exactly |
| `!=` | Value must not match exactly |
| `~` | Value must be present in record |
| `!~` | Value must not be present in record |

For example, for this input record:

```
id=12345 msg="it worked"
```

The following filters match:

```
id=12345
id!=4567
msg~worked
msg!~failed
```

The key must always be present to a match to occur; none of the following filters match
the above record:

```
tag=test
tag!=test
tag~test
tag!~test
```

### Formatting

Format strings are templates that replace the placeholder values, which must start with
a `.`, be valid logfmt keys, and must follow a character that could not be part of a valid
logfmt key, with the values of the like-named key.

For instance, for the following record:

```
id=12345 msg="it worked"
```

The following format strings would produce the given output:

| format | output |
| --- | --- |
| `.id` | 12345 |
| `[.id] .msg` | [12345] it worked |
| `.id.msg` | 12345.msg |
| `.tag (.id): .msg |  (12345): it worked |


### Which Is Provided?

When `lf` receives args that include a `|`, that character is expected to divide
filters from formats; the output format is to the right of the last `|`, and all else
is a filter (additional `|` characters are redundant).

If `lf` receives no `|` in its args, it uses hueristics to decide if it was given a
filter or a format.  When in doubt, the args are assumed to be a format.
