import * as React from 'react'
import {ReactElement} from 'react'
import {RouteComponentProps} from 'react-router'
import {APIStore, LayoutStore, RouterStore} from 'app/stores'
import {STORE_API, STORE_LAYOUT, STORE_ROUTER} from 'app/constants'

export interface PageProps extends RouteComponentProps<any> {
  [ STORE_LAYOUT ]?: LayoutStore
  [ STORE_ROUTER ]?: RouterStore
  [ STORE_API ]?: APIStore
}

export abstract class Page<S> extends React.Component<PageProps, S> {
  title: string
  layout: LayoutStore = this.props[ STORE_LAYOUT ]
  router: RouterStore = this.props[ STORE_ROUTER ]
  api: APIStore = this.props[ STORE_API ]

  componentWillMount (): void {
    this.layout.setFullScreen(false)
    this.layout.title = this.renderTitle()
  }

  protected renderTitle (): ReactElement<any> {
    return (
      <div className='pt-navbar-heading'>
        {this.title}
      </div>
    )
  }
}

export abstract class FullScreenPage<S> extends Page<S> {

  componentWillMount (): void {
    this.layout.setFullScreen(true)
  }
}