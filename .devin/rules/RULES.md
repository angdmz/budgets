---
trigger: manual
description: 
globs: 
---

# execution guideline
 - all project execution has to happen through docker-compose.yml, it is the entrypoint for the system manipulation


# bug fixing
every bug fix must include first reproducing it, planning a fix, write tests that cover the bug, then fix it, if the bug happens in both UI and backend, there must be test cases in both and an integration selenium test case in tests/ folder

# general coding
we should not have fallback logic any way, unless indicated on prompt, let's favor a failfast generally speaking