import * as React from 'react'
import {inject, observer} from "mobx-react";
import {Page} from "app/components";
import {STORES} from "app/constants";
import {Classes, Switch} from "@blueprintjs/core";

@inject(...STORES)
@observer
export class Settings extends Page<{}> {
  title = 'Settings'

  private onDarkSwitched = () => {
    this.layout.theme(!this.layout.dark)
  }

  render () {
    return (
      <div className='page'>
        <div>
          <label className={Classes.LABEL}>Theme</label>
          <Switch
            checked={this.layout.dark}
            onChange={this.onDarkSwitched}
            labelElement={<strong>Dark</strong>}/>
        </div>
      </div>
    )
  }
}