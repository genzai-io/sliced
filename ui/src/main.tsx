import * as React from 'react'
import * as ReactDOM from 'react-dom'
import {Main} from 'app'

// render react DOM
ReactDOM.render(
  React.createElement('div', {}, <Main/>),
  document.getElementById('root')
)
