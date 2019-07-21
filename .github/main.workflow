workflow "New workflow" {
  on = "push"
  resolves = ["make"]
}

action "make" {
  uses = "make"
}
