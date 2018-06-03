import * as React from 'react'
import {inject, observer} from "mobx-react";
import {Page} from "app/components";
import {STORES} from "app/constants";

@inject(...STORES)
@observer
export class Home extends Page<{}> {
  render () {
    return (
      <div className='page'>
        <button className='pt-button pt-intent-success' onClick={() => {
          this.layout.theme(!this.layout.dark)
        }}>Theme
        </button>

        <button className='pt-button pt-intent-danger' onClick={() => {
          this.layout.setFullScreen(!this.layout.fullScreen)
        }}>Full-Screen
        </button>

        <button className='pt-button' onClick={() => {
          this.router.push('/auth')
        }}>Auth
        </button>
      </div>
    )
  }
}