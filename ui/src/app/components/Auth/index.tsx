import * as React from 'react'
import {FullScreenPage} from "app/components";
import {STORES} from "app/constants";
import {inject, observer} from "mobx-react";

import {AutoSizer} from 'react-virtualized'

@inject(...STORES)
@observer
export class Auth extends FullScreenPage<{}> {
  render () {
    return (
      <AutoSizer>
        {({ height, width }) => (
          <div style={{ width: width + 'px', height: height + 'px', padding: '30px', backgroundColor: 'red' }}>

          </div>
        )}
      </AutoSizer>
    )
  }
}