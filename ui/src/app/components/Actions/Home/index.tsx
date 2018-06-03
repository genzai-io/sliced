import * as React from 'react'
import {inject, observer} from "mobx-react";
import {Page} from "app/components";
import {STORES} from "app/constants";

@inject(...STORES)
@observer
export class Actions extends Page<{}> {
  title = 'Actions'

  // protected get title () {
  //   return 'Actions'
  // }

  render () {
    return (
      <div className='page'>

      </div>
    )
  }
}