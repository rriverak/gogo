import { Component, Host, h, Prop, Event, EventEmitter } from '@stencil/core';

@Component({
  tag: 'gogo-video-modal',
  styleUrl: 'gogo-video-modal.css',
  shadow: true,
})
export class GogoVideoModal {

  @Prop() header: string;
  @Event() close: EventEmitter;
  closeHandler() {
    this.close.emit();
  }

  render() {
    return (
      <Host>
        <div class="modal">
          <div class="modal-overlay"></div>
          <div class="modal-container">
            <div class="modal-content">
              <div class="modal-title">
                {this.header}
              </div>
              <div class="body">
                <slot></slot>
              </div>
              <div class="modal-footer">
                <button type="button" onClick={() => { this.closeHandler() }} class="indigo">Ok</button>
              </div>
            </div>
          </div>
        </div>
      </Host>
    );
  }

}
