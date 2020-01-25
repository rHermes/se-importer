# se-importer


## What
This tool imports data from the XML files found at
[StackExchange data dumps](https://archive.org/download/stackexchange/)
into an SQL Server database.

**It is still very much work in progress.**

## How
You will have to download the 7zip files and put them in a folder. You then
run this tool pointing it at that folder. For now it picks up the MSSQL connection
details from environment variables.

The files being 7z poses a bit of a challenge. I found a pure go library, but
the performance was bad. Found another library which was a wrapper for the the
C library, but couldn't get valid data when reading with it. So ended up with a module
which is really just a wrapper for the `7z` executable. I think that I will make it so
that on windows, it uses the pure go library and on linux it uses the wrapper, since it's
so much faster.

It uses the `encoding/xml` package in the stdlib to decode the XML. Only thing of note
here is that `time.Time` doesn't implement unmarshaling for XML, so I have to implement a
dummy type called `SEDate` and use that.

I've made the decision to keep all the sites in the same tables, and instead have a
`site` table, which I use in all the other tables. This makes cross site queries much easier
and also makes the tables bigger, which is nice for practice as performance problems
pop up earlier.

## Why
I love SQL and I have a fair share of experience with Postgres, the best
database <3, but I wanted to expand my horizons and learn more about
SQL Server and T-SQL. I needed some data to do so and when I googled
for "big public datasets", I got the StackExchange datadump files.

Found some tools to import it into SQL Server, but they where mostly for
windows. Decided I might as well just get started with learning how to
use MSSQL from Go, since that is my language of choice.
