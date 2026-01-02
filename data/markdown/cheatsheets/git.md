# git

## Basics

Initialize repo:  
`git init`

Clone repo:  
`git clone <url>`

Show status:  
`git status -sb`

Stage files:  
`git add <file>`

Stage all:  
`git add -A`

Commit:  
`git commit -m "message"`

Commit all tracked:  
`git commit -am "message"`

Commit without hooks:  
`git commit --no-verify -m "message"`

## Branches

List branches:  
`git branch -vv`

Create branch:  
`git branch <name>`

Create and switch:  
`git switch -c <name>`

Switch branch:  
`git switch <name>`

Delete branch:  
`git branch -d <name>`

Force delete branch:  
`git branch -D <name>`

Rename branch:  
`git branch -m <new-name>`

## Remotes

List remotes:  
`git remote -v`

Add remote:  
`git remote add <name> <url>`

Fetch:  
`git fetch --all --prune`

Pull (rebase):  
`git pull --rebase`

Push:  
`git push <remote> <branch>`

Set upstream on first push:  
`git push -u <remote> <branch>`

## Log and inspect

One-line log graph:  
`git log --oneline --graph --decorate --all`

Show a commit:  
`git show <commit>`

Show file history:  
`git log -- <file>`

Who last changed a line:  
`git blame <file>`

## Diff

Unstaged diff:  
`git diff`

Staged diff:  
`git diff --cached`

Diff against branch:  
`git diff <branch>`

Word diff:  
`git diff --word-diff`

## Undo and recover

Unstage file:  
`git restore --staged <file>`

Discard working changes:  
`git restore <file>`

Restore file from commit:  
`git restore --source <commit> <file>`

Amend last commit:  
`git commit --amend`

Revert a commit:  
`git revert <commit>`

Reset branch (soft/mixed/hard):  
`git reset --soft <commit>`  
`git reset --mixed <commit>`  
`git reset --hard <commit>`

Find lost commits:  
`git reflog`

## Stash

Stash changes:  
`git stash push -m "msg"`

Stash including untracked:  
`git stash -u`

List stashes:  
`git stash list`

Apply latest stash:  
`git stash apply`

Pop latest stash:  
`git stash pop`

Stash only staged:  
`git stash push --staged -m "msg"`

## Rebase

Interactive rebase:  
`git rebase -i <upstream>`

Continue/abort rebase:  
`git rebase --continue`  
`git rebase --abort`

Rebase onto new base:  
`git rebase --onto <newbase> <oldbase> <branch>`

## Fixup and autosquash

Create fixup commit:  
`git commit --fixup <commit>`

Interactive rebase with autosquash:  
`git rebase -i --autosquash <upstream>`

Fixup by message match:  
`git commit --fixup=amend:<commit>`

## Cherry-pick

Cherry-pick a commit:  
`git cherry-pick <commit>`

Continue/abort cherry-pick:  
`git cherry-pick --continue`  
`git cherry-pick --abort`

## Tags

List tags:  
`git tag`

Create annotated tag:  
`git tag -a <tag> -m "msg"`

Push tags:  
`git push --tags`

## Merge

Merge branch:  
`git merge <branch>`

Abort merge:  
`git merge --abort`

## Bisect

Start bisect:  
`git bisect start`

Mark good/bad:  
`git bisect good <commit>`  
`git bisect bad <commit>`

End bisect:  
`git bisect reset`

## Worktrees

Add worktree:  
`git worktree add <path> <branch>`

List worktrees:  
`git worktree list`

Remove worktree:  
`git worktree remove <path>`

## Submodules

Add submodule:  
`git submodule add <url> <path>`

Init/update submodules:  
`git submodule update --init --recursive`

Update to latest:  
`git submodule update --remote`

## Config

Set name/email:  
`git config --global user.name "Name"`  
`git config --global user.email "name@example.com"`

Enable rerere:  
`git config --global rerere.enabled true`

Show config:  
`git config --list`
