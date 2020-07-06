import { Component, Host, h, State, Prop, Watch } from '@stencil/core';

@Component({
  tag: 'gogo-video-sidebar',
  styleUrl: 'gogo-video-sidebar.css',
  shadow: true,
})
export class GogoVideoSidebar {
  @Prop() state:any;
  @State() activeSection: string = "Users";
  @State() users = [];

  componentDidLoad() {
    if(this.state){
      this.stateChanged(this.state)
    }
  }

  @Watch("state")
  stateChanged(newState){
    this.users = newState.Users;
  }

  render() {
    return (
      <Host>
        <ul>
          <li>
            <a class={this.activeSection == "Users" ? "active" : null} onClick={() => { this.activeSection = "Users" }} >Users</a>
          </li>
          <li>
            <a class={this.activeSection == "Chat" ? "active" : null} onClick={() => { this.activeSection = "Chat" }}  >Chat</a>
          </li>
        </ul>
        {
          this.activeSection == "Users" ?
            <gogo-video-sidebar-section name="Users">
              <table>
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Name</th>
                  </tr>
                </thead>
                <tbody>
                  {this.users.map((row) => {
                    return <tr>
                      <td>{row.ID}</td>
                      <td>{row.Name}</td>
                    </tr>
                  })}
                  {this.users.length == 0 ? <tr><td class="center" colSpan={2}>No Data</td></tr> : null}
                </tbody>
              </table>
            </gogo-video-sidebar-section> : null
        }

      </Host>
    );
  }

}
