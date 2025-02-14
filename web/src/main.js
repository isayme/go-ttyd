import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { Terminal } from '@xterm/xterm';
import debounce from 'debounce';
import { pack } from 'msgpackr';

// organize-imports-ignore 
import 'normalize.css';
// organize-imports-ignore 
import '@xterm/xterm/css/xterm.css';

import './style.css';

const appElement = document.getElementById('app')

const wsUrl = window.location.protocol + "//" + window.location.host + '/ws';
// const wsUrl = window.location.protocol + "//" + window.location.host.replace('5173', '1323') + '/ws';
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

const BUF_DATA_PREFIX = new Uint8Array([0])
const BUF_RESIZE_PREFIX = new Uint8Array([1])

function mergeUint8Array(buf1, buf2) {
  let buf = new Uint8Array(buf1.length + buf2.length)
  buf.set(buf1, 0)
  buf.set(buf2, buf1.length)
  return buf
}

const handleOnData = function (data) {
  let buf = mergeUint8Array(BUF_DATA_PREFIX, new TextEncoder().encode(data))
  websocket.send(buf)
}

terminal.onData(handleOnData)
terminal.onBinary(handleOnData)

terminal.onResize((evt) => {
  let rect = appElement.getBoundingClientRect()

  let buf = mergeUint8Array(BUF_RESIZE_PREFIX, pack({
    cols: evt.cols,
    rows: evt.rows,
    x: rect.width,
    y: rect.height,
  }))

  websocket.send(buf)
})

const onResizeHandler = debounce(function () {
  fitAddon.fit()
}, 100)

// 监听窗口resize事件，调整terminal窗口
window.addEventListener('resize', onResizeHandler)

// 调整页面标题
terminal.onTitleChange((title) => {
  document.title = title
})

// 绑定到html元素
terminal.open(appElement)


websocket.onopen = function () {
  console.log('websocket connect ok')

  // 自适应大小
  fitAddon.fit()
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