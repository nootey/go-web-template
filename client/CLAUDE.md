# Template - Client Instructions

## Stack

Vue 3 + TypeScript, PrimeVue 4 components, PrimeFlex utility classes.

## Styling

- Do NOT write custom CSS classes. Use inline PrimeFlex utility classes instead.
- For responsive styles that require a media query, use a scoped `<style>` block with an `id`-based selector - never a custom class.
- Do not use arbitrary inline `style` attributes for spacing/layout - reach for PrimeFlex first.

## Components

- Use PrimeVue components where one exists for the use case before writing a custom component.
- Check related existing pages/components for examples.

## Validation

- Every form field must use existing methods of validation (see any _Form_ component for reference)
- Success, error and validation messages come from the backend; just relay them in the client via toast_service.

## Code Style Guidelines

- TypeScript: Strict type checking, ES modules, explicit return types
- Use relative paths only - no aliases like `@/` or `~/`
- Use `import type` for type-only imports; group them at the end of the import block
