import * as React from 'react'
import {inject, observer} from "mobx-react";
import {Page} from "app/components";
import {STORES} from "app/constants";

@inject(...STORES)
@observer
export class Dashboard extends Page<{}> {
  title = 'Dashboard'

  render () {
    return (
      <div className='page'>

      </div>
    )
  }
}