# vim

## Basics

Open a file:  
`vim <file>`

Save:  
`:w`

Save and quit:  
`:wq`

Quit:  
`:q`

Force quit (discard changes):  
`:q!`

Quit with error:  
`:cq`
Useful to abort an accidental git commit for example.

Help for a command:  
`:help <command>`

## Modes

Insert mode:  
`i` insert at cursor  
`a` append after cursor  
`o` open new line below  
`O` open new line above

Normal mode:  
`Esc`

Visual mode (char/line/block):  
`v` / `V` / `Ctrl-v`

Command-line mode:  
`:` in normal mode

## Movement

Left/down/up/right:  
`h` `j` `k` `l`

Next/previous word:  
`w` / `b`

End of word:  
`e`

Start/end of line:  
`0` / `^` / `$`

Top/bottom of file:  
`gg` / `G`

Go to line number:  
`:<line-number>`

## Editing

Delete character:  
`x`

Delete line:  
`dd`

Yank (copy) line:  
`yy`

Paste after/before:  
`p` / `P`

Undo/redo:  
`u` / `Ctrl-r`

Change word:  
`cw`

Change inner word:  
`ciw`

Replace character:  
`r<character>`

Join lines:  
`J`

Repeat last change:  
`.`

## Search

Search forward/backward:  
`/pattern` / `?pattern`

Next/previous match:  
`n` / `N`

Search word under cursor:  
`*` / `#`

Toggle search highlight:  
`:set hlsearch` / `:set nohlsearch`

Clear current highlights:  
`:noh`

## Replace

Replace in file:  
`:%s/old/new/g`

Replace with confirmation:  
`:%s/old/new/gc`

Replace in selection:  
`:'<,'>s/old/new/g`

## Visual mode

Yank/delete selection:  
`y` / `d`

Indent/outdent selection:  
`>` / `<`

## Files and buffers

Edit a file:  
`:e <file>`

Write to a new file:  
`:w <file>`

List buffers:  
`:ls`

Next/previous buffer:  
`:bnext` / `:bprev`

Delete buffer:  
`:bd`

## Windows and splits

Horizontal/vertical split:  
`:sp` / `:vs`

Move between windows:  
`Ctrl-w w` or `Ctrl-w h/j/k/l`

Close current/other windows:  
`Ctrl-w c` / `Ctrl-w o`

Equalize window sizes:  
`Ctrl-w =`

## Tabs

New/next/prev tab:  
`:tabnew` / `:tabnext` / `:tabprev`

Close tab:  
`:tabclose`

## Registers and clipboard

List registers:  
`:reg`

Default register (unnamed):  
`yy` then `p`

Named register (a-z):  
`"ayy` then `"ap`

Append to named register (A-Z):  
`"Ayy` then `"Ap`

Last yank register (0):  
`"0p`

Numbered delete registers (1-9):  
`"1p`

Black hole register (discard):  
`"_d`

Yank to system clipboard:  
`"+y`

Paste from system clipboard:  
`"+p`

## Macros

Start/stop recording:  
`q<register>` / `q`

Play macro:  
`@<register>`

Repeat last macro:  
`@@`

## Settings

Show line numbers:  
`:set number`

Show relative numbers:  
`:set relativenumber`

Ignore case unless uppercase:  
`:set ignorecase smartcase`

Toggle wrap:  
`:set wrap` / `:set nowrap`
