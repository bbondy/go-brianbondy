# GDB

## Execution flow:

Start the program:  
`run`  
`run cli args here`


Stop the program being debugged:  
`kill`  
`k`

Go to the next line of code:  
`next`  
`next <offset>` (next that many times)  
`n`

Step inside the function, or next line of code:  
`step`  
`step <offset>` (step that many times)  
`s`

Resume execution:  
`continue`  
`continue <ignore_count>` (number of breakpoints to ignore in a row)  
`c`

Step out of current function (specifically out of current stack frame, recursion):  
`finish`  
`fin`

Step out of current loop (above current mem address):  
`until`  
`u`

Execute until something is hit:  
`until <line_num>`  
`until <relative_file_path>:<line_num>`  
`until <some_func_name>`  
`until <relative_file_path>:<some_func_name>`

## Breakpoints:

Set a breakpoint:  
`break <line_num>`  
`break <relative_file_path>:<line_num>`  
`break <some_func_name>`  
`break <relative_file_path>:<some_func_name>`  
`b <line_num>`

Clear a breakpoint:  
`clear <line_num>`  
`clear <relative_file_path>:<line_num>`  
`clear <some_func_name>`  
`clear <relative_file_path>:<some_func_name>`  
`clear` (currently stopped at location's breakpoint)  
`cl <line_num>`

Set a breakpoint by regex:  
`rbreak <regex>`  
`rb <regex>`

Set a conditional breakpoint:  
`break <line_num> if z > 10`  
`break <relative_file_path>:<some_func_name> if z > 10`  
`break <relative_file_path>:<some_func_name> if strlen(str) === 0 && z > 0`

Set a temporary (one time) breakpoint:  
`tbreak <line_num>`  
`tb <line_num>`
`enable once <breakpoint_num>` (disables instead of deleting breakpoint)

Add a condition to an existing breakpoint:  
`condition <breakpoint_num> my_var == 7` (where 1 is the breakpoint number)  
`cond <breakpoint_num> my_var == 7`

Set a watchpoint on a variable:  
`watch z`  
`w z`

Set a conditional watchpoint:  
`watch (z > 10)`

Get a list of breakpoints, watchpoints, and catchpoints:  
`info break`  
`i b`

Get a list of watchpoints:  
`info watch`  
`i w`

Disable a breakpoint:  
`disable <breakpoint_num>`
`disable <breakpoint_num_list>` (space separated)  
`disable` (all breakpoints)

Enable a breakpoint:  
`enable <breakpoint_num>`
`enable <breakpoint_num_list>` (space separated)  
`enable` (all breakpoints)

Delete a breakpoint, watchpoint, or catchpoint:  
`delete <breakpoint_num>`  
`delete <breakpoint_num_list>` (space separated)  
`delete` (all breakpoints)  
`d <breakpoint_num>`

Breakpoint command lists:
```
commands <breakpoint_num>
# silent optionally supresses output from the debugger.
silent
printf "test %d", i
# continue optionally continues the program again after the commands are done.
continue
end
```


## Inspecting

Printing a value of a variable:  
`print val`  
`p val`

Backtrace:  
`backtrace`  
`bt`


Looking at a stack frame in the callstack:  
`frame 0` (for first frame)  
`frame 1` (for the thing that called that, etc)

Go up to the parent frame in the callstack:  
`up`

Go down the callstack to the next child frame:  
`down`

## Custom commands

Defining custom commands:
```
define <custom_command_name>
...
end
```

## Command history

Last command:  
`Ctrl+P`  
`<Up arrow>` if not in tui mode

Next command:  
`Ctrl+N`  
`<Down arrow>` if not in tui mode

Repeat last command:  
`<enter>`

## GDB control

TUI mode via command line:  
`gdb -tui` (via command line)

Enter / exit TUI mode:  
`Ctrl+X+A`

Startup file:  
`~/.gdbinit` (for general purpose things)  
`$PWD/.gdbinit` (for per project config)  
`~/.gdbinit 
`gdb -command=alt-gdb-init-file a.out` (Start gdb for program a.out with gdb init file gdb-init-file)

Quit GDB:  
`quit`  
`q`

## Source navigation:

Change region of code used for displaying:  
`list sourcefile.cc:1`  
`l sourcefile.cc:1`


## Disassembly:

`g++ -S <source_file>`  (produce assembly language file)
