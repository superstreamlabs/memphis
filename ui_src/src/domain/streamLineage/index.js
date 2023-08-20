// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useContext, useState, useRef } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { useHistory } from 'react-router-dom';

import { convertBytes, parsingDate } from '../../services/valueConvertor';
import { BsZoomIn, BsZoomOut } from 'react-icons/bs';
import { MdZoomOutMap } from 'react-icons/md';
import { IoCloseOutline } from 'react-icons/io5';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { Canvas, Node, Edge, Label } from 'reaflow';
import { httpRequest } from '../../services/http';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import { Divider, message } from 'antd';
import pathDomains from '../../router';
import Connection from './components/connection';
import Station from './components/station';

const fake_data = {
    stations: [
        {
            id: 1,
            name: 'a',
            dls_messages: 90,
            total_messages: 132006
        },
        {
            id: 8,
            name: 'a1',
            dls_messages: 0,
            total_messages: 145
        },
        {
            id: 9,
            name: 'a2',
            dls_messages: 0,
            total_messages: 145
        },
        {
            id: 10,
            name: 'a3',
            dls_messages: 0,
            total_messages: 145
        },
        {
            id: 11,
            name: 'a4',
            dls_messages: 0,
            total_messages: 145
        },
        {
            id: 12,
            name: 'a5',
            dls_messages: 0,
            total_messages: 145
        },
        {
            id: 13,
            name: 'a6',
            dls_messages: 0,
            total_messages: 145
        }
    ],
    apps: [
        {
            app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
            consumers: [],
            producers: [
                {
                    name: 'a',
                    station_id: 1,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a1',
                    station_id: 8,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a2',
                    station_id: 9,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a3',
                    station_id: 10,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                }
            ],
            from: [],
            to: [1, 8, 9, 10]
        },
        {
            app_id: '93919a6c-d550-40a8-9ca3-51365496768b',
            consumers: [
                {
                    name: 'c022',
                    station_id: 10,
                    app_id: '93919a6c-d550-40a8-9ca3-51365496768b',
                    count: 1
                },
                {
                    name: 'c0232',
                    station_id: 9,
                    app_id: '93919a6c-d550-40a8-9ca3-51365496768b',
                    count: 1
                },
                {
                    name: 'c9020202',
                    station_id: 8,
                    app_id: '93919a6c-d550-40a8-9ca3-51365496768b',
                    count: 2
                }
            ],
            producers: [],
            from: [1, 8, 9, 10],
            to: [11, 12]
        },
        {
            app_id: 'e9727623-0b9f-49ec-be3c-66593184b99c',
            consumers: [],
            producers: [
                {
                    name: 'a',
                    station_id: 1,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a1',
                    station_id: 8,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a2',
                    station_id: 9,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a3',
                    station_id: 10,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                }
            ],
            from: [],
            to: [11, 12, 13]
        },
        {
            app_id: 'e9727623-0b9f-49ec-be3c-66593184b349c',
            consumers: [],
            producers: [
                {
                    name: 'a',
                    station_id: 1,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a1',
                    station_id: 8,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a2',
                    station_id: 9,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                },
                {
                    name: 'a3',
                    station_id: 10,
                    app_id: 'e9727623-0b9f-49ec-be3c-66593184b16c',
                    count: 1
                }
            ],
            from: [11, 12, 13],
            to: []
        }
    ]
};

const StreamLineage = ({ expend, setExpended }) => {
    const [state, dispatch] = useContext(Context);
    const url = window.location.href;
    const [isLoading, setisLoading] = useState(false);
    const [processing, setProcessing] = useState(false);
    const [messageData, setMessageData] = useState({});
    const [nodes, setNodes] = useState([]);
    const [edges, setEdges] = useState([]);
    const [zoom, setZoom] = useState(0.7);
    const ref = useRef(null);

    const history = useHistory();

    const getPosionMessageDetails = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_GRAPH_OVERVIEW);
            arrangeData(fake_data);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'overview' });
        getPosionMessageDetails();
    }, []);

    // useEfrfect(() => {
    //     let sub;
    //     const jc = JSONCodec();
    //     const sc = StringCodec();

    //     const subscribeAndListen = async (subName, pubName, messageId) => {
    //         try {
    //             const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.${subName}.${messageId}`, sc.encode('SUB'));

    //             if (rawBrokerName) {
    //                 const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
    //                 sub = state.socket?.subscribe(`$memphis_ws_pubs.${pubName}.${messageId}.${brokerName}`);
    //                 listenForUpdates(subName, messageId);
    //             }
    //         } catch (err) {
    //             console.error(`Error subscribing to ${subName} data for messageId ${messageId}:`, err);
    //             return;
    //         }
    //     };
    //     const listenForUpdates = async (subName, messageId) => {
    //         try {
    //             if (sub) {
    //                 for await (const msg of sub) {
    //                     let data = jc.decode(msg.data);
    //                     arrangeData(data);
    //                 }
    //             }
    //         } catch (err) {
    //             console.error(`Error receiving ${subName} data updates for messageId ${messageId}:`, err);
    //         }
    //     };

    //     subscribeAndListen('poison_message_journey_data', 'poison_message_journey_data', messageId);

    //     return () => {
    //         if (sub) {
    //             try {
    //                 sub.unsubscribe();
    //             } catch (err) {
    //                 console.error('Error unsubscribing from message journey data:', err);
    //             }
    //         }
    //     };
    // }, [state.socket]);

    const arrangeData = (data) => {
        let nodesList = [];
        let edgesList = [];
        if (data) {
            data['stations']?.map((row, index) => {
                let node = {
                    id: row.id,
                    text: row.name,
                    width: 300,
                    height: 300,
                    data: {
                        value: 'station',
                        name: row.name,
                        dls_messages: row.dls_messages,
                        total_messages: row.total_messages
                    }
                };
                nodesList.push(node);
            });
            data['apps']?.map((row, index) => {
                let node = {
                    id: row.app_id,
                    text: 'app',
                    width: 300,
                    height: 260,
                    data: {
                        value: 'app',
                        producers: row.producers,
                        consumers: row.consumers
                    }
                };
                row.from.map((from, index) => {
                    let edge = {
                        id: `${row.app_id}-${from}`,
                        from: from,
                        to: row.app_id,
                        fromPort: from,
                        toPort: row.app_id,
                        selectionDisabled: true
                    };
                    edgesList.push(edge);
                });
                row.to.map((to, index) => {
                    let edge = {
                        id: `${row.app_id}-${to}`,
                        from: row.app_id,
                        to: to,
                        fromPort: row.app_id,
                        toPort: to,
                        selectionDisabled: true
                    };
                    edgesList.push(edge);
                });
                nodesList.push(node);
            });

            setEdges(edgesList);
            setNodes(nodesList);
            setTimeout(() => {
                setisLoading(false);
            }, 1000);
        }
    };

    return (
        <div className={expend ? 'stream-lineage-container' : 'stream-lineage-container overview-components-wrapper lineage-smaller'}>
            <div className="title-wrapper">
                <div className="bread-crumbs">
                    <p>Graph View</p>
                </div>
                <div className="actions-wrapper">
                    <div className="close-wrapper">
                        {expend && <IoCloseOutline onClick={() => setExpended(false)} />}
                        {!expend && <MdZoomOutMap onClick={() => setExpended(true)} />}
                    </div>
                    <div className="zoom-wrapper">
                        <BsZoomIn onClick={() => ref.current.zoomIn()} />
                        <Divider />
                        <BsZoomOut onClick={() => ref.current.zoomOut()} />
                        <Divider />
                        <span className="fit-wrapper" onClick={() => ref.current.fitCanvas()}>
                            Fit
                        </span>
                    </div>
                </div>
            </div>
            {isLoading && (
                <div className="loader-uploading">
                    <Loader background={false} auto={false} />
                </div>
            )}
            {!isLoading && (
                <div className="canvas-wrapper">
                    <Canvas
                        className="canvas"
                        readonly={true}
                        direction="RIGHT"
                        nodes={nodes}
                        edges={edges}
                        fit={true}
                        ref={ref}
                        zoom={zoom}
                        maxZoom={0.2}
                        minZoom={-0.9}
                        height={'100%'}
                        maxHeight={nodes?.length < 5 ? 700 : nodes?.length * 170}
                        node={
                            <Node style={{ stroke: 'transparent', fill: 'transparent', strokeWidth: 1 }} label={<Label style={{ display: 'none' }} />}>
                                {(event) => (
                                    <foreignObject height={event.height} width={event.width} x={0} y={0} className="node-wrapper">
                                        {event.node.data.value === 'app' && (
                                            <Connection id={event.node.id} producers={event.node.data.producers} consumers={event.node.data.consumers} />
                                        )}
                                        {event.node.data.value === 'station' && (
                                            <Station
                                                stationName={event.node.data.name}
                                                dls_messages={event.node.data.dls_messages}
                                                total_messages={event.node.data.total_messages}
                                            />
                                        )}
                                    </foreignObject>
                                )}
                            </Node>
                        }
                        zoomable={true}
                        arrow={null}
                        edge={(edge) => <Edge {...edge} className={'edge'} />}
                    />
                </div>
            )}
        </div>
    );
};
export default StreamLineage;
