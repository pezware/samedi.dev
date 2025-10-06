# License Information

## Overview

Samedi is licensed under the **MIT License**.

This is one of the most permissive open source licenses, allowing you to:
- ✅ Use the software commercially
- ✅ Modify the software
- ✅ Distribute the software
- ✅ Sublicense the software
- ✅ Use it privately

The only requirements are:
- Include the license and copyright notice in any copy of the software/substantial portions

## Files

### LICENSE
The main MIT License file at the root of the repository.

**Copyright**: 2025 Samedi Contributors

### .license-header.txt
Template for license headers in Go source files.

## Adding License Headers

### New Go Files

All new Go source files should include this header:

```go
// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package yourpackage
```

**Why SPDX?**
- SPDX (Software Package Data Exchange) is a standard format
- Enables automated license scanning and compliance tools
- Widely recognized in the open source community

### Example

```go
// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package plan

import (
	"context"
	"fmt"
)

// Manager handles learning plan operations.
type Manager struct {
	// ...
}
```

### Non-Go Files

For documentation, scripts, and other files, license headers are optional but recommended:

**Markdown**:
```markdown
<!--
Copyright (c) 2025 Samedi Contributors
SPDX-License-Identifier: MIT
-->
```

**Shell scripts**:
```bash
#!/bin/bash
# Copyright (c) 2025 Samedi Contributors
# SPDX-License-Identifier: MIT
```

**YAML/TOML**:
```yaml
# Copyright (c) 2025 Samedi Contributors
# SPDX-License-Identifier: MIT
```

## Contributing

### Your Contributions

By contributing to Samedi, you agree that your contributions will be licensed under the MIT License.

This means:
- Your code will be open source under MIT
- Others can use, modify, and distribute it
- You retain copyright to your contributions
- You're granting broad permissions to use your work

See [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

### Third-Party Dependencies

When adding dependencies:
1. **Check the license** - Ensure it's compatible with MIT
2. **Compatible licenses**: MIT, BSD, Apache 2.0, ISC
3. **Incompatible licenses**: GPL, AGPL (copyleft licenses)

**Check dependency licenses**:
```bash
go-licenses csv github.com/pezware/samedi.dev/cmd/samedi
```

### Preferred Dependencies

Choose dependencies with permissive licenses:
- MIT ✅
- BSD (2-clause, 3-clause) ✅
- Apache 2.0 ✅
- ISC ✅
- Unlicense ✅
- Public Domain ✅

Avoid:
- GPL, LGPL, AGPL ❌ (copyleft - requires derivative works to use same license)
- Proprietary ❌
- Unknown ❌

## License Compatibility

MIT is compatible with:
- ✅ Proprietary software (can be used in closed-source products)
- ✅ Other permissive licenses (BSD, Apache 2.0)
- ✅ GPL (MIT code can be included in GPL projects)

MIT is NOT compatible as a *whole* with:
- ❌ GPL if you want to keep your project MIT (GPL is viral)

**Best practice**: Keep Samedi MIT, don't mix with GPL dependencies.

## Copyright Notice

### Why "Samedi Contributors"?

Using "Samedi Contributors" instead of individual names:
- Simplifies copyright management
- Acknowledges all contributors collectively
- Standard practice for community-driven projects (like Linux, Kubernetes)

### Individual Attribution

Individual contributors are recognized in:
- `CONTRIBUTORS.md` - List of all contributors
- Git commit history - Your commits are permanently attributed to you
- Release notes - Major contributions are highlighted

## Automated License Checking

### Pre-commit Hook

The project includes a pre-commit hook that checks for:
- Missing license headers in new Go files
- Incompatible third-party licenses

### CI/CD

GitHub Actions workflow checks:
- All Go files have proper license headers
- Dependencies use compatible licenses
- No GPL/AGPL dependencies

## FAQ

### Do I need to update the year?

**Short answer**: No need to update yearly.

**Long answer**: The year (2025) represents when the project was created. Some projects update it to a range (e.g., "2025-2026"), but it's not required by the MIT License.

### Can I use Samedi in a commercial product?

**Yes!** MIT License allows commercial use without restrictions.

### Do I need to share my modifications?

**No.** MIT doesn't require you to share your changes (unlike GPL).

However, we'd love if you contributed improvements back to the project! 🙏

### What if I fork Samedi?

You can:
- Keep the MIT License (recommended)
- Change to a more restrictive license (not recommended)
- Add additional terms (not recommended)

You must:
- Keep the original MIT License text
- Keep the copyright notice

### Can I remove the license headers?

**No.** The license headers are required by the MIT License terms:

> "The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software."

## Tools

### Check License Headers

```bash
# Find Go files without license headers
find . -name "*.go" -type f ! -path "*/vendor/*" -exec grep -L "SPDX-License-Identifier: MIT" {} \;
```

### Add License Headers Automatically

```bash
# Using addlicense tool
go install github.com/google/addlicense@latest

addlicense -c "Samedi Contributors" -l mit -s .go .
```

### Verify Dependency Licenses

```bash
# Using go-licenses
go install github.com/google/go-licenses@latest

go-licenses check github.com/pezware/samedi.dev/cmd/samedi
go-licenses report github.com/pezware/samedi.dev/cmd/samedi
```

## Resources

- **MIT License**: https://opensource.org/licenses/MIT
- **SPDX Identifiers**: https://spdx.org/licenses/
- **Choose a License**: https://choosealicense.com/
- **Open Source Guide**: https://opensource.guide/legal/

## Contact

Questions about licensing?
- Open a [Discussion](https://github.com/pezware/samedi.dev/discussions)
- Contact @arbeitandy
- Read the [LICENSE](../LICENSE) file

---

**TL;DR**: Samedi uses MIT License. Add headers to new Go files. You can use it for anything. Contributions are MIT-licensed.
