// Package uuid provides an implementation of *Universally Unique Identifiers*
// that can intelligently reorder their bytes, such that V1 (timestamp-based)
// UUIDs are stored in a "dense" byte order.
//
// This "dense" format has the property that UUIDs which were generated in
// increasing chronological order will also sort in increasing lexical order.
// This property makes them suitable for use in an indexed column of a SQL
// database, as modern SQL database implementations are much happier if their
// keys are inserted in lexically increasing order.
//
// For more on the byte-reordering technique:
// https://www.percona.com/blog/2014/12/19/store-uuid-optimized-way/
//
// Additionally, instead of using the binary format (hard to read) OR the
// traditional 8-4-4-4-12 text format (too verbose), we offer a new
// SQL-friendly text format of "@<base64>".  This is about halfway in length
// between the two standard representations, but is human-readable and can be
// copy-pasted in a terminal.
//
// A concrete example: the UUID `f8a2571a-8add-11e8-96a8-185e0fad6335` is
// reordered to "dense" format as `[11 e8 8a dd f8 a2 57 1a 96 a8 ...]`, and this
// "dense" binary representation is serialized to "@<base64>" format as
// `@EeiK3fiiVxqWqBheD61jNQ`.  (Note that all bytes from `96 a8` onward have
// the same order in both binary representations.)
//
package uuid
