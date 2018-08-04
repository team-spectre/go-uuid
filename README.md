# go-uuid

A database-friendly implementation of Universally Unique Identifiers for Go.

[![License](https://img.shields.io/github/license/chronos-tachyon/go-uuid.svg?maxAge=86400)](https://github.com/chronos-tachyon/go-uuid/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/chronos-tachyon/go-uuid?status.svg)](https://godoc.org/github.com/chronos-tachyon/go-uuid)
[![Build Status](https://img.shields.io/travis/chronos-tachyon/go-uuid.svg?maxAge=3600&logo=travis)](https://travis-ci.org/chronos-tachyon/go-uuid)
[![Coverage Status](https://img.shields.io/coveralls/github/chronos-tachyon/go-uuid.svg?maxAge=3600&logo=travis)](https://coveralls.io/github/chronos-tachyon/go-uuid)
[![Issues](https://img.shields.io/github/issues/chronos-tachyon/go-uuid.svg?maxAge=7200&logo=github)](https://github.com/chronos-tachyon/go-uuid/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/chronos-tachyon/go-uuid.svg?maxAge=7200&logo=github)](https://github.com/chronos-tachyon/go-uuid/pulls)
[![Latest Release](https://img.shields.io/github/release/chronos-tachyon/go-uuid.svg?maxAge=2592000&logo=github)](https://github.com/chronos-tachyon/go-uuid/releases)

## Quick Start

```
import "github.com/chronos-tachyon/go-uuid"

// Generate a new version 1 UUID
u := uuid.New()

// Print the UUID as a string like "77b99cea-8ab4-11e8-96a8-185e0fad6335".
fmt.Println(u.CanonicalString())

// Print the UUID as a string like "@EeiKtHe5nOqWqBheD61jNQ".
fmt.Println(u.DenseString())

// Serialize to SQL.  Defaults to a string like "@EeiKtHe5nOqWqBheD61jNQ",
// but you can also serialize as a canonical UUID string (plus several
// variants thereof), a lexical-order BLOB, or even a standard-order BLOB,
// depending on u.SetPreferences.
db.Exec(`INSERT INTO table (uuid) VALUES (?)`, u)

// Choose serialization preferences for this UUID object.
u.SetPreferences(uuid.Preferences{
  // Pick whether u.Value() returns text or a BLOB.
  Value:  uuid.Binary,

  // Pick whether BLOB serializations use lexical or standard byte ordering.
  // Also selects BLOBs deserialization behavior, as it can make guesses.
  Binary: uuid.DenseFirst,

  // Pick which format to use for text serializations.
  // Does not affect text deserialization; all formats are always recognized.
  Text:   uuid.Dense,
})

// Deserialize from SQL.
row := db.QueryRow(`SELECT uuid FROM table LIMIT 1`)
row.Scan(&u)
```

## What's it about?

This is a fully-featured UUID library in Golang, comparable to the excellent
[github.com/satori/go.uuid](https://github.com/satori/go.uuid).  However, one
thing that I found lacking about satori's package is that, like all other UUID
implementations that I've found for Go, it isn't very friendly for use in SQL
databases.

"But chronos, satori's package supports sql.Scanner and driver.Valuer!"

Ah yes, but the standard *byte order* for UUIDs, codified in RFC 4122, is
a poor choice for use in indexed columns. In fact, if you try to use a
standard UUID as a `PRIMARY KEY`, you're going to tank your database
performance and spend years wondering what happened.

The fundamental issue is that the standard byte order for UUIDs puts the
fields in the wrong order, particularly if you're using Version 1 (time-based)
UUIDs.

Here's a quick anatomy lesson of a V1 UUID:

```
          timestamp = aa bb cc dd ee ff gg hh (60 bits plus 4-bit version)
    sequence number =                   ii jj (14 bits plus 2-bit variant)
        MAC address =       kk ll mm nn oo pp (48 bits)

    +-------------+-------+-------+-------+-------------------+
    | ee ff gg hh | cc dd | aa bb | ii jj | kk ll mm nn oo pp |
    +-------------+-------+-------+-------+-------------------+
```

Notice how the timestamp gets scrambled? That hurts SQL performance because
SQL indexes are happy when keys are strictly increasing and very, very sad
when they are not.  The ideal would be to have the timestamp in pure
big-endian order, followed by the sequence number in big-endian order, then
finally the MAC address.

(You could argue that "MAC address then sequence number" makes *slightly* more
sense, but it's kind of a wash either way, and it's fairly rare for two UUIDs
generated on two different machines to share the same 100ns clock tick.)

After we reorder the bytes for SQL, the new UUID anatomy looks like this:

```
             out[0] = in[6]
             out[1] = in[7]
             out[2] = in[4]
             out[3] = in[5]
             out[4] = in[0]
             out[5] = in[1]
             out[6] = in[2]
             out[7] = in[3]
                    |
      ______________|____________
     /                           \
    +-------+-------+-------------+-------+-------------------+
    | aa bb | cc dd | ee ff gg hh | ii jj | kk ll mm nn oo pp |
    +-------+-------+-------------+-------+-------------------+
```

### Base-64 representation

That's all well and good, but what if we don't want a BLOB? Suppose we require
that our PRIMARY KEY be some string that can be copy-pasted from a terminal.
What's a good textual representation for a UUID that uses this lexical byte
order?

Well, encoding binary data in a textual format is kind of a solved problem.
Why not use base-64? It already provides a pure-ASCII encoding with only 33%
blowup, much better than the 100+% blowup of the canonical UUID text format.
And if we prepend some distinctive ASCII character from outside the base-64
repertoire, such as "@", then we get a very quick way of determining which
parser to use. So we move the bytes around to big-endian lexical order, then
write a single "@", then write the base-64 representation of the lexical
bytes. The length is implicit (a UUID is exactly 16 bytes), so we don't need
any padding characters.

Examples:

    331a90da-9013-11e8-aad2-42010a800002 => @EeiQEzMakNqq0kIBCoAAAg
    335984e8-9013-11e8-aad2-42010a800002 => @EeiQEzNZhOiq0kIBCoAAAg

Result: 36 characters becomes 23, for a 56% savings, AND the new text format
has the property that two temporally-ordered V1 UUIDs will sort correctly
after text conversion.

### Automatic inference of byte order

As it turns out, there's *just* enough structure in a UUID that we can make an
educated guess about which order a BLOB was serialized in:

- The version field (top 4 bits of "aa") is always a value between 0001 and
  0101, i.e. 5 values out of 15. That gives a 31.25% chance of a false
  positive.

- The variant field (top 2 bits of "ii") is always 10xx xxxx. At 1 value out
  of 4, that gives a 25% chance of a false positive.

- The probabilities are independent, so they combine as 7.8125%.

- Timestamps before 1970-01-01 or after the current date plus 5 years are
  soft-rejected as implausible. The probability of landing in that timestamp
  range by chance will grow over time, but by 2100-01-01 it will only be
  0.2309%. That is also an independent probability, so the odds of reaching
  this point by chance are 0.0180%.

- In the extremely rare event that both interpretations have correct version
  bits, correct variant bits, AND plausible timestamps, the timestamp is
  reimagined as a probability by assuming a normal distribution with the
  current time as mean and one century as standard deviation.  The most
  probable interpretation is then preferred.
