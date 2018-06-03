import * as React from 'react'
import {inject, observer} from 'mobx-react'
import {Page} from 'app/components'
import {STORES} from 'app/constants'

import FontAwesomeIcon from '@fortawesome/react-fontawesome'
import {faAws} from '@fortawesome/fontawesome-free-brands'


@inject(...STORES)
@observer
export class Cluster extends Page<{}> {
  renderTitle () {
    return <FontAwesomeIcon icon={faAws} size={'3x'}/>
  }

  render () {
    return (
      <div className='page'>
        {/*<h2>Cluster</h2>*/}
      </div>
    )
  }
}