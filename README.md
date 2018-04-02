1.创建socketAcceptor-->SessionManager-->newSocketPeer-->HandlerChainManagerImplement
                                                     -->SetReadWriteChain   
2.启动socketAcceptor-->socketSession-->readChain&writeChain-->readChain.Call(ev)
                                                           -->ev.ChainSend.Call(ev)
                                                           -->writeChain.Call(ev)

1.创建socketConnector-->SessionManager-->socketConnector-->newSocketPeer-->HandlerChainManagerImplement
                                                                        -->SetReadWriteChain  
2.启动socketConnector-->socketSession-->readChain&writeChain-->readChain.Call(ev)
                                                           -->ev.ChainSend.Call(ev)
                                                           -->writeChain.Call(ev)
