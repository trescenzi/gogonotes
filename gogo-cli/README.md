# GoGo CLI

A simple interface to gogonotes it enables you to, `download`, `save`, and,
create `new` notes.

## Required Environment Vars

- `$HASURA_ADMIN_SECRET`: the admin secret to be used in api requests. Set to the
    header `x-hasura-admin-secret`
- `$HASURA_ENDPOINT`: the url of the hasura api you're using to access notes.

## Nonrequired Environment Vars

- `$GOGONOTES_ROOT`: the folder within which to sync notes. Defaults to
    `$HOME/notes`.
- `$EDITOR`: used within the `new` command to control which editor the new note
    is opened within.

## Commands

### download

`gogonotes download` will sync all notes from the db to `$GOGONOTES_ROOT`. These
notes will be named `<id>.ggn`.

#### Example

```
$ ls $GOGONOTES_ROOT

$ gogonotes download
$ ls $GOGONOTES_ROOT
1.ggn    104.ggn  11.ggn   13.ggn   15.ggn   17.ggn   19.ggn   20.ggn   22.ggn   24.ggn   26.ggn   28.ggn   3.ggn    31.ggn   33.ggn   35.ggn   37.ggn   39.ggn   40.ggn   42.ggn   44.ggn   46.ggn   48.ggn   5.ggn    51.ggn   7.ggn    83.ggn   85.ggn   87.ggn   89.ggn   90.ggn
10.ggn   105.ggn  12.ggn   14.ggn   16.ggn   18.ggn   2.ggn    21.ggn   23.ggn   25.ggn   27.ggn   29.ggn   30.ggn   32.ggn   34.ggn   36.ggn   38.ggn   4.ggn    41.ggn   43.ggn   45.ggn   47.ggn   49.ggn   50.ggn   6.ggn    8.ggn    84.ggn   86.ggn   88.ggn   9.ggn
```

### save <note-name>

`gogonotes save <note-name>` will save the relevant note to the db. `$GOGONOTES_ROOT` is assumed
and `.ggn` is as well.

#### Example

```
$ gogonotes save 16
Saving $GOGONOTES_ROOT/16.ggn
Success! Updated Note 16

```

### new <name?>

`gogonotes new <name?>` will first download notes then find the next id and create a
new note with that id, and if provided, name. Then it will open `$EDITOR` for
editing. Upon closing the editor the note will be saved.

#### Example

```
$ gogonotes new note
// editor
Saving $GOGONOTES_ROOT/note-17.ggn
Success! Updated Note 17
```
