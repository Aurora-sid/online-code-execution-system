---
description: Reviews code changes for bugs, style issues, and best practices. Use when reviewing PRs or checking code quality.
---

# Code Review Skill

When reviewing code, follow these steps:

## Review checklist

1. **Correctness**: Does the code do what it's supposed to?
2. **Edge cases**: Are error conditions handled?
3. **Style**: Does it follow project conventions?
4. **Performance**: Are there obvious inefficiencies?
5. **Architecture & Structure ** ：Is the code organized in a scalable and maintainable way?Separation of Concerns: Does the code follow the project's architectural pattern (e.g., MVC, Clean Architecture)? Is logic placed in the correctlayer(Controller vs Service vs Repository)?Modularity: Are components cohesive? Is there unnecessary coupling between modules?
Dependency Management: Are dependencies clearly defined? Is there any circular dependency introduced?

## How to provide feedback

- Be specific about what needs to change
- Explain why, not just what
- Suggest alternatives when possible