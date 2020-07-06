# gogo-video



<!-- Auto Generated Below -->


## Properties

| Property    | Attribute    | Description | Type     | Default     |
| ----------- | ------------ | ----------- | -------- | ----------- |
| `SessionId` | `session-id` |             | `string` | `undefined` |


## Dependencies

### Depends on

- [gogo-video-device-selection](../gogo-video-device-selection)
- [gogo-video-debug](../gogo-video-debug)
- [gogo-video-sidebar](../gogo-video-sidebar)

### Graph
```mermaid
graph TD;
  gogo-video --> gogo-video-device-selection
  gogo-video --> gogo-video-debug
  gogo-video --> gogo-video-sidebar
  gogo-video-device-selection --> gogo-video-modal
  gogo-video-sidebar --> gogo-video-sidebar-section
  style gogo-video fill:#f9f,stroke:#333,stroke-width:4px
```

----------------------------------------------

*Built with [StencilJS](https://stenciljs.com/)*
