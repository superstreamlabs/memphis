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
import { BsZoomIn, BsZoomOut } from 'react-icons/bs';
import { Canvas, Node, Edge, Label } from 'reaflow';
import { IoRefresh } from 'react-icons/io5';
import { StringCodec, JSONCodec } from 'nats.ws';
import { MdZoomOutMap } from 'react-icons/md';
import { IoClose } from 'react-icons/io5';
import { Divider } from 'antd';

import { ApiEndpoints } from '../../const/apiEndpoints';
import graphPlaceholder from '../../assets/images/graphPlaceholder.svg';
import BackIcon from '../../assets/images/backIcon.svg';
import { httpRequest } from '../../services/http';
import Connection from './components/connection';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Station from './components/station';
import Button from '../../components/button';
import LockFeature from '../../components/lockFeature';

const StreamLineage = ({ expend, setExpended, createStationTrigger }) => {
    const [state, dispatch] = useContext(Context);
    const [isLoading, setisLoading] = useState(false);
    const [nodes, setNodes] = useState([]);
    const [edges, setEdges] = useState([]);
    const ref = useRef(null);

    const getGraphData = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_GRAPH_OVERVIEW);
            arrangeData(data);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        getGraphData();
    }, []);

    useEffect(() => {
        const sc = StringCodec();
        const jc = JSONCodec();
        let sub;

        const subscribeToOverviewData = async () => {
            try {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.get_graph_overview`, sc.encode('SUB'));

                if (rawBrokerName) {
                    const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                    sub = state.socket?.subscribe(`$memphis_ws_pubs.get_graph_overview.${brokerName}`);
                    listenForUpdates();
                }
            } catch (err) {
                console.error('Error subscribing to overview data:', err);
            }
        };

        const listenForUpdates = async () => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        arrangeData(data);
                    }
                }
            } catch (err) {
                console.error('Error receiving graph data updates:', err);
            }
        };

        expend && subscribeToOverviewData();

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from graph data:', err);
                }
            }
        };
    }, [state.socket, expend]);

    const arrangeData = (data) => {
        let nodesList = [];
        let edgesList = [];
        if (data) {
            data['stations']?.map((row, index) => {
                let node = {
                    id: row.id,
                    text: row.name,
                    width: 300,
                    height: row.schema_name !== '' ? 280 : 250,
                    data: {
                        value: 'station',
                        name: row.name,
                        dls_messages: row.dls_messages,
                        total_messages: row.total_messages,
                        schema_name: row.schema_name
                    },
                    ports: [
                        {
                            id: `${row.id}_east`,
                            side: 'EAST',
                            width: 10,
                            height: 10,
                            hidden: true
                        },
                        {
                            id: `${row.id}_west`,
                            side: 'WEST',
                            width: 10,
                            height: 10,
                            hidden: true
                        }
                    ]
                };
                nodesList.push(node);
            });
            const sortedArray = data['apps']?.slice().sort((a, b) => {
                return a.app_id.localeCompare(b.app_id);
            });
            sortedArray?.map((row, index) => {
                let node = {
                    id: row.app_id,
                    text: 'app',
                    width: 300,
                    height: 100 + row.producers.length * 30 + row.consumers.length * 30,
                    data: {
                        value: 'app',
                        producers: row.producers,
                        consumers: row.consumers
                    }
                };
                row.from.map((from, index) => {
                    let edge = {
                        id: `${row.app_id}-${from.station_id}`,
                        from: from.station_id,
                        to: row.app_id,
                        fromPort: `${from.station_id}_east`,
                        toPort: row.app_id,
                        selectionDisabled: true,
                        data: {
                            active: from.active
                        }
                    };
                    edgesList.push(edge);
                });
                row.to.map((to, index) => {
                    let edge = {
                        id: `${row.app_id}-${to.station_id}`,
                        from: row.app_id,
                        to: to.station_id,
                        fromPort: row.app_id,
                        toPort: `${to.station_id}_west`,
                        selectionDisabled: true,
                        data: {
                            active: to.active
                        }
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
        <div
            className={
                expend
                    ? 'stream-lineage-container'
                    : nodes?.length > 0
                    ? 'stream-lineage-container overview-components-wrapper lineage-smaller'
                    : 'overview-components-wrapper lineage-empthy'
            }
        >
            <div className="title-wrapper">
                <div className="overview-components-header">
                    <p>System overview</p>
                    <label>A dynamic, self-built graph visualization of your main system components</label>
                </div>
                {!expend && nodes?.length > 0 && (
                    <div className="refresh-wrapper" onClick={() => getGraphData()}>
                        <IoRefresh />
                    </div>
                )}
                {nodes?.length > 0 && (
                    <div className="actions-wrapper">
                        <div
                            className="close-wrapper"
                            onClick={() =>
                                expend
                                    ? setExpended(false)
                                    : state?.userData?.entitlements && state?.userData?.entitlements['feature-graph-overview']
                                    ? setExpended(true)
                                    : null
                            }
                        >
                            {expend && <IoClose />}

                            {!expend && <MdZoomOutMap />}
                            {!expend && state?.userData?.entitlements && !state?.userData?.entitlements['feature-graph-overview'] && (
                                <LockFeature header={'Full screen'} />
                            )}
                        </div>
                        {expend && (
                            <div className="zoom-wrapper">
                                <BsZoomIn onClick={() => ref.current.zoomIn()} />
                                <Divider />
                                <BsZoomOut onClick={() => ref.current.zoomOut()} />
                                <Divider />
                                <span className="fit-wrapper" onClick={() => ref.current.fitCanvas()}>
                                    Fit
                                </span>
                            </div>
                        )}
                    </div>
                )}
            </div>
            {isLoading && (
                <div className="loader-uploading">
                    <Loader background={false} auto={false} />
                </div>
            )}
            {!isLoading && nodes?.length > 0 && (
                <div className="canvas-wrapper">
                    <Canvas
                        className="canvas"
                        readonly={true}
                        direction="RIGHT"
                        nodes={nodes}
                        edges={edges}
                        fit={true}
                        ref={ref}
                        zoomable={state?.userData?.entitlements && state?.userData?.entitlements['feature-graph-overview'] ? true : false}
                        maxZoom={0.2}
                        minZoom={-0.9}
                        height={'100%'}
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
                                                schema_name={event.node.data.schema_name}
                                            />
                                        )}
                                    </foreignObject>
                                )}
                            </Node>
                        }
                        arrow={null}
                        edge={(edge) => <Edge {...edge} className={edge?.data?.active === true ? 'edge processing' : 'edge'} />}
                    />
                </div>
            )}
            {!isLoading && nodes?.length === 0 && (
                <div className="empty-connections-container">
                    <img src={graphPlaceholder} alt="graphPlaceholder" onClick={() => createStationTrigger(true)} />
                    <p>There are no entities to display</p>
                    <span className="desc">Please create at least one entity, such as a station, to display the graph overview.</span>
                    <Button
                        className="modal-btn"
                        height="34px"
                        placeholder={'Start by create a new station'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        onClick={() => createStationTrigger(true)}
                    />
                </div>
            )}
        </div>
    );
};
export default StreamLineage;
