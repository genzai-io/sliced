import * as React from 'react'
import {inject, observer} from "mobx-react";
import {Page} from "app/components";
import {STORES} from "app/constants";

@inject(...STORES)
@observer
export class Alerts extends Page<{}> {
  title = 'Alerts'

  render () {
    return (
      <div className='page'>

      </div>
    )
  }
}