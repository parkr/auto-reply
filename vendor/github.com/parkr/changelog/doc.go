// changelog provides a means of parsing and printing markdown changelogs.
//
// The basic structure of a changelog is a reverse-chronological list of
// versions and the changes they contain. Each version can have its own
// direct list of uncategorized changes, and can contain a set of
// subsections. Subsections are a means of categorizing sets of changes
// based on component or type of change. Each change consists of a summary
// and a reference – either a pull request or issue number, or a @mention
// to the contributing user.
//
// A basic changelog might look something like:
//
//     ## 1.0.0 / 2015-02-21
//
//     ### Major Enhancemens
//
//       * Added that big feature (#1425)
//
//     ### Bug Fixes
//
//       * Fixed that narsty bug with tokenization (@carla)
//
//     ## 0.0.1 / 2015-02-20
//
//       * Initial implementation
//       * Tokenize a changelog (#1)
package changelog
