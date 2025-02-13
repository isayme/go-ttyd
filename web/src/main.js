import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { Terminal } from '@xterm/xterm';
// organize-imports-ignore 
import 'normalize.css';
// organize-imports-ignore 
import '@xterm/xterm/css/xterm.css';

import './style.css';

const appElement = document.getElementById('app')
if (!appElement) {
  throw new Error(`element #app not found`)
}

// const wsUrl = window.location.protocol + "//" + window.location.host + '/ws';
const wsUrl = window.location.protocol + "//" + window.location.host.replace('5173', '1323') + '/ws';
const websocket = new WebSocket(wsUrl)
websocket.binaryType = 'arraybuffer'

const terminal = new Terminal({
  theme: {
    background: 'black',
    foreground: 'white',
  },
})

const fitAddon = new FitAddon()
terminal.loadAddon(fitAddon)
terminal.loadAddon(new WebLinksAddon());

// 调整页面标题
terminal.onTitleChange((title) => {
  document.title = title
})

// 绑定到html元素
terminal.open(appElement)

// 自适应大小
fitAddon.fit()

function onResizeHandler() {
  fitAddon.fit()
}

terminal.onData(function (data) {
  websocket.send(new TextEncoder().encode('\x00' + data))
})

terminal.onBinary(function (data) {
  websocket.send(new TextEncoder().encode('\x00' + data))
})

terminal.onResize((evt) => {
  let rect = appElement.getBoundingClientRect()
  let resizeMessage = {
    cols: evt.cols,
    rows: evt.rows,
    x: rect.width,
    y: rect.height,
  }
  websocket.send(
    new TextEncoder().encode('\x01' + JSON.stringify(resizeMessage)),
  )
})

// 监听窗口resize事件，调整terminal窗口
window.addEventListener('resize', onResizeHandler)

websocket.onopen = function () {
  console.log('ws open')
}

websocket.onmessage = function (evt) {
  let data = evt.data
  terminal.write(typeof data === 'string' ? data : new Uint8Array(data))
}

websocket.onclose = function () {
  terminal.write('\n\rconnection closed')
}

websocket.onerror = function (a, ev) {
  terminal.write('\n\rconnection error')
}
