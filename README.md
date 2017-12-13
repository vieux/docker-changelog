# Installation

```
$ go get github.com/vieux/docker-changelog
```

# Usage

```
docker-changelog "<version> (YYYY-MM-DD)" <from_branch>..<to_branch>
```

# Example

```
docker-changelog "17.12.0-ce (2017-12-05)" refs/heads/17.11..refs/heads/master > CHANGELOG.md
```

```
docker-changelog "17.11.0-ce (2017-11-04)" refs/heads/17.10..refs/heads/17.11 > CHANGELOG.md
```