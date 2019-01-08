# csved

An editor for plaintext tables, consisting of columns separated by arbitrary
characters and rows separated by newlines, i.e. tsv, csv and related formats.

Currently it's quite spartan, but more is planned - see below.

## usage

It will assume comma-separated data for files ending in `.csv`, and assume
tab-separated data for files ending in `.tsv` - both case-insensitive. Otherwise
it will prompt you to choose a delimiter for the data. This can be anything
except a newline, which will break the parser for obvious reasons. Newlines in
the data will also cause strange behaviour (row breaks where there shouldn't be
any) which is a natural consequence of a line-by-line parser.

You can run it with the flag `-d` for extra debugging information.

### keys

- Arrows, Emacs or vi movement keys - move around
- `RET` - Edit the currently selected cell
- `C-t` - toggle treating the first row as a title for the column
- `C-s` - save the file (truncating it if it already exists)
- `C-c` - quit

## plans

- Add rows - `C-r` (add **R**ow)
- Del rows - `C-u` ("Emacs" binding)
- Add columns - `C-l` (add co**L**umn)
- Del columns - `C-k` (Emacs binding for kill)
- Search - `C-f` (CUA)
- Clipboard
- Home and End, tab (excel-esque navigation)
- Tab completion in a column

## copying

Licensed MIT.
