import * as React from 'react'
import {ReactElement} from 'react'
import {hot} from 'react-hot-loader'
import {Route, Router, Switch} from 'react-router'
import {Provider} from 'mobx-react'
// import createBrowserHistory from 'history/createBrowserHistory'
import createHashHistory from 'history/createHashHistory'
import {createStores} from 'app/stores'
import {Layout} from 'app/layout'
import {IPageRouteProps} from 'app/route'

import './style.css'
// import 'normalize.css/normalize.css'
// import '@blueprintjs/core/lib/css/blueprint.css'
// import '@blueprintjs/icons/lib/css/blueprint-icons.css'
// Most of react-virtualized's styles are functional (eg position, size).
// Functional styles are applied directly to DOM elements.
// The Table component ships with a few presentational styles as well.
// They are optional, but if you want them you will need to also import the CSS file.
// This only needs to be done once probably during your application's bootstrapping process.
import 'react-virtualized/styles.css'

// prepare MobX stores
const history = createHashHistory()//createBrowserHistory()
const rootStore = createStores(history)

/**
 *
 */
class LayoutRoute extends Route<IPageRouteProps> {
  render () {
    const { page, ...props } = this.props
    return React.createElement(page, { ...props })
  }
}

/**
 *
 */
export class Root extends React.Component<any, any> {
  devtools: ReactElement<any>

  componentWillMount () {
    // if (process.env.NODE_ENV !== 'production') {
    //   const DevTools = require('mobx-react-devtools').default
    //   this.devtools = <DevTools/>
    // }
  }

  render () {
    if (this.devtools && this.devtools != null) {
      return (
        <div>
          {this.props.children}
          {this.devtools}
        </div>
      )
    } else {
      return this.props.children
    }
  }
}

// render react DOM
const App = hot(module)(({ scopes, props, history }) => {
  return (
    <Root {...props}>
      <Router history={history}>
        <Switch>
          <LayoutRoute path='/' page={Layout}/>
        </Switch>
      </Router>
    </Root>
  )
})

export class Main extends React.Component<any, any> {
  render () {
    return (
      <Provider {...rootStore}>
        <App history={history}/>
      </Provider>
    )
  }
}
