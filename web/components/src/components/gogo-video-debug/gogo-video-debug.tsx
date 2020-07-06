import { Component, Host, h } from '@stencil/core';

@Component({
  tag: 'gogo-video-debug',
  styleUrl: 'gogo-video-debug.css',
  shadow: true,
})
export class GogoVideoDebug {

  render() {
    return (
      <Host>
        <table>
          <tbody>
            <tr>
              <td>test</td>
            </tr>
          </tbody>
        </table>
      </Host>
    );
  }

}
