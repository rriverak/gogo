/* eslint-disable */
/* tslint:disable */
/**
 * This is an autogenerated file created by the Stencil compiler.
 * It contains typing information for all components that exist in this project.
 */
import { HTMLStencilElement, JSXBase } from "@stencil/core/internal";
export namespace Components {
    interface GogoVideo {
        "SessionId": string;
    }
    interface GogoVideoDebug {
    }
    interface GogoVideoDeviceSelection {
    }
    interface GogoVideoModal {
        "header": string;
    }
    interface GogoVideoSidebar {
        "state": any;
    }
    interface GogoVideoSidebarSection {
        "name": string;
    }
}
declare global {
    interface HTMLGogoVideoElement extends Components.GogoVideo, HTMLStencilElement {
    }
    var HTMLGogoVideoElement: {
        prototype: HTMLGogoVideoElement;
        new (): HTMLGogoVideoElement;
    };
    interface HTMLGogoVideoDebugElement extends Components.GogoVideoDebug, HTMLStencilElement {
    }
    var HTMLGogoVideoDebugElement: {
        prototype: HTMLGogoVideoDebugElement;
        new (): HTMLGogoVideoDebugElement;
    };
    interface HTMLGogoVideoDeviceSelectionElement extends Components.GogoVideoDeviceSelection, HTMLStencilElement {
    }
    var HTMLGogoVideoDeviceSelectionElement: {
        prototype: HTMLGogoVideoDeviceSelectionElement;
        new (): HTMLGogoVideoDeviceSelectionElement;
    };
    interface HTMLGogoVideoModalElement extends Components.GogoVideoModal, HTMLStencilElement {
    }
    var HTMLGogoVideoModalElement: {
        prototype: HTMLGogoVideoModalElement;
        new (): HTMLGogoVideoModalElement;
    };
    interface HTMLGogoVideoSidebarElement extends Components.GogoVideoSidebar, HTMLStencilElement {
    }
    var HTMLGogoVideoSidebarElement: {
        prototype: HTMLGogoVideoSidebarElement;
        new (): HTMLGogoVideoSidebarElement;
    };
    interface HTMLGogoVideoSidebarSectionElement extends Components.GogoVideoSidebarSection, HTMLStencilElement {
    }
    var HTMLGogoVideoSidebarSectionElement: {
        prototype: HTMLGogoVideoSidebarSectionElement;
        new (): HTMLGogoVideoSidebarSectionElement;
    };
    interface HTMLElementTagNameMap {
        "gogo-video": HTMLGogoVideoElement;
        "gogo-video-debug": HTMLGogoVideoDebugElement;
        "gogo-video-device-selection": HTMLGogoVideoDeviceSelectionElement;
        "gogo-video-modal": HTMLGogoVideoModalElement;
        "gogo-video-sidebar": HTMLGogoVideoSidebarElement;
        "gogo-video-sidebar-section": HTMLGogoVideoSidebarSectionElement;
    }
}
declare namespace LocalJSX {
    interface GogoVideo {
        "SessionId"?: string;
    }
    interface GogoVideoDebug {
    }
    interface GogoVideoDeviceSelection {
        "onClose"?: (event: CustomEvent<any>) => void;
    }
    interface GogoVideoModal {
        "header"?: string;
        "onClose"?: (event: CustomEvent<any>) => void;
    }
    interface GogoVideoSidebar {
        "state"?: any;
    }
    interface GogoVideoSidebarSection {
        "name"?: string;
    }
    interface IntrinsicElements {
        "gogo-video": GogoVideo;
        "gogo-video-debug": GogoVideoDebug;
        "gogo-video-device-selection": GogoVideoDeviceSelection;
        "gogo-video-modal": GogoVideoModal;
        "gogo-video-sidebar": GogoVideoSidebar;
        "gogo-video-sidebar-section": GogoVideoSidebarSection;
    }
}
export { LocalJSX as JSX };
declare module "@stencil/core" {
    export namespace JSX {
        interface IntrinsicElements {
            "gogo-video": LocalJSX.GogoVideo & JSXBase.HTMLAttributes<HTMLGogoVideoElement>;
            "gogo-video-debug": LocalJSX.GogoVideoDebug & JSXBase.HTMLAttributes<HTMLGogoVideoDebugElement>;
            "gogo-video-device-selection": LocalJSX.GogoVideoDeviceSelection & JSXBase.HTMLAttributes<HTMLGogoVideoDeviceSelectionElement>;
            "gogo-video-modal": LocalJSX.GogoVideoModal & JSXBase.HTMLAttributes<HTMLGogoVideoModalElement>;
            "gogo-video-sidebar": LocalJSX.GogoVideoSidebar & JSXBase.HTMLAttributes<HTMLGogoVideoSidebarElement>;
            "gogo-video-sidebar-section": LocalJSX.GogoVideoSidebarSection & JSXBase.HTMLAttributes<HTMLGogoVideoSidebarSectionElement>;
        }
    }
}
