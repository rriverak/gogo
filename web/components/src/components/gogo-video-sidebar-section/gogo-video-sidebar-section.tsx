import { Component, Host, h, Prop } from '@stencil/core';

@Component({
  tag: 'gogo-video-sidebar-section',
  styleUrl: 'gogo-video-sidebar-section.css',
  shadow: true,
})
export class GogoVideoSidebarSection {
  @Prop() name: string;
  render() {
    return (
      <Host>
        <section>
          <div class="body">
            <slot></slot>
          </div>
        </section>
      </Host>
    );
  }

}
