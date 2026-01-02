# css

## Basics

Link a stylesheet:  
`<link rel="stylesheet" href="styles.css">`

Inline style:  
`<div style="color: red;">`

Basic rule:  
```css
selector {
  property: value;
}
```

## Selectors

Class/ID/type:  
`.btn` / `#header` / `button`

Descendant/child:  
`.card p` / `.card > p`

Sibling:  
`h2 + p` / `h2 ~ p`

Attribute:  
`input[type="email"]`

Pseudo-class:  
`:hover` `:focus` `:nth-child(2n)`

Pseudo-element:  
`::before` `::after`

## Specificity and cascade

Specificity order:  
`inline > id > class/attr/pseudo-class > element`

Important:  
`color: red !important;`

Reduce conflicts with layers:  
```css
@layer reset, base, components, utilities;
@layer components {
  .btn { padding: 8px 12px; }
}
```

## Box model

Padding/border/margin:  
`padding: 8px;` `border: 1px solid #000;` `margin: 16px;`

Border box sizing:  
`box-sizing: border-box;`

## Layout: flex

Create flex container:  
`display: flex;`

Direction and wrap:  
`flex-direction: row;` `flex-wrap: wrap;`

Align items/content:  
`align-items: center;` `justify-content: space-between;`

Item sizing:  
`flex: 1 1 200px;`

## Layout: grid

Create grid:  
`display: grid;`

Columns/rows:  
`grid-template-columns: repeat(3, 1fr);`  
`grid-template-rows: auto 1fr auto;`

Gap:  
`gap: 16px;`

Place item:  
`grid-column: 1 / 3;` `grid-row: 2;`

## Positioning

Static/relative/absolute:  
`position: static;` `relative;` `absolute;`

Fixed/sticky:  
`position: fixed;` `position: sticky;` `top: 0;`

## Typography

Font stack:  
`font-family: "IBM Plex Sans", system-ui, sans-serif;`

Font size/line height:  
`font-size: 16px;` `line-height: 1.5;`

Text overflow:  
`white-space: nowrap;` `text-overflow: ellipsis;` `overflow: hidden;`

## Color and background

Color formats:  
`#09f` `rgb(0 153 255)` `hsl(210 100% 50%)`

Gradient background:  
`background: linear-gradient(135deg, #111, #444);`

## Effects

Shadow:  
`box-shadow: 0 8px 24px rgb(0 0 0 / 0.2);`

Filters:  
`filter: blur(6px) saturate(1.2);`

## Transforms and transitions

Transform:  
`transform: translateY(-4px) scale(1.02);`

Transition:  
`transition: transform 200ms ease, opacity 200ms ease;`

## Animation

Keyframes:  
```css
@keyframes float {
  0% { transform: translateY(0); }
  100% { transform: translateY(-6px); }
}
```

Use animation:  
`animation: float 2s ease-in-out infinite alternate;`

## Media queries

Responsive breakpoints:  
```css
@media (min-width: 768px) {
  .grid { grid-template-columns: 1fr 1fr; }
}
```

## Container queries

Define container:  
`container-type: inline-size;`

Query by container:  
```css
@container (min-width: 480px) {
  .card { display: grid; }
}
```

## Custom properties (variables)

Define and use:  
```css
:root { --gap: 12px; }
.stack { gap: var(--gap); }
```

Fallback:  
`color: var(--accent, #0ea5e9);`

## Functions

Calc/clamp/min/max:  
`width: calc(100% - 24px);`  
`font-size: clamp(14px, 2vw, 20px);`  
`height: min(60vh, 480px);`

## Advanced selectors

Matches/negation:  
`:is(.primary, .ghost)` / `:not(.disabled)`

Has parent state:  
`:has(> img)`  

## Accessibility

Focus-visible:  
`:focus-visible { outline: 2px solid #0ea5e9; }`

Reduce motion:  
```css
@media (prefers-reduced-motion: reduce) {
  * { animation: none; transition: none; }
}
```

## Debugging

Outline everything:  
`* { outline: 1px solid red; }`

Check layout overflow:  
`body { overflow: hidden; }`
