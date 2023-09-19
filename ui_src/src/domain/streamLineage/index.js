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
import { MdZoomOutMap } from 'react-icons/md';
import { IoClose } from 'react-icons/io5';
import { Divider } from 'antd';

import graphPlaceholder from '../../assets/images/graphPlaceholder.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { ReactComponent as GraphPlaceholder } from '../../assets/images/graphPlaceholder.svg';
import BackIcon from '../../assets/images/backIcon.svg';
import LockFeature from '../../components/lockFeature';
import { httpRequest } from '../../services/http';
import Connection from './components/connection';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Station from './components/station';

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
            const sortedProducers = data['producers']?.slice().sort((a, b) => {
                return a.app_id.localeCompare(b.app_id);
            });
            const sortedConsumers = data['consumers']?.slice().sort((a, b) => {
                return a.app_id.localeCompare(b.app_id);
            });
            sortedProducers?.map((producer, index) => {
                let node = {
                    id: `${producer.name}-${producer.station_id}`,
                    text: 'producer',
                    width: 200,
                    height: 100,
                    data: {
                        value: 'producer',
                        producer_details: producer
                    }
                };
                let edge = {
                    id: `${producer.name}-${producer.station_id}`,
                    from: `${producer.name}-${producer.station_id}`,
                    to: producer.station_id,
                    toPort: `${producer.station_id}_west`,
                    selectionDisabled: true,
                    data: {
                        active: true,
                        is_producer: true
                    }
                };
                nodesList.push(node);
                edgesList.push(edge);
            });
            sortedConsumers?.map((consumer, index) => {
                let node = {
                    id: `${consumer.name}-${consumer.station_id}`,
                    text: 'consumer',
                    width: 200,
                    height: 100,
                    data: {
                        value: 'consumer',
                        consumer_details: consumer
                    }
                };
                let edge = {
                    id: `${consumer.name}-${consumer.station_id}`,
                    from: consumer.station_id,
                    to: `${consumer.name}-${consumer.station_id}`,
                    fromPort: `${consumer.station_id}_east`,
                    selectionDisabled: true,
                    data: {
                        active: true,
                        is_producer: false
                    }
                };
                nodesList.push(node);
                edgesList.push(edge);
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
                    <div className="flex">
                        {expend && <img src={BackIcon} onClick={() => setExpended(false)} alt="backIcon" />}

                        <p>System overview</p>
                    </div>
                    <label>A dynamic, self-built graph visualization of your main system components</label>
                </div>
                {nodes?.length > 0 && (
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
                        maxZoom={0.1}
                        minZoom={-0.9}
                        maxHeight={nodes.length > 3 ? nodes.length * 350 : 900}
                        node={
                            <Node style={{ stroke: 'transparent', fill: 'transparent', strokeWidth: 1 }} label={<Label style={{ display: 'none' }} />}>
                                {(event) => (
                                    <foreignObject height={event.height} width={event.width} x={0} y={0} className="node-wrapper">
                                        {event.node.data.value === 'producer' && <Connection id={event.node.id} producer={event.node.data.producer_details} />}
                                        {event.node.data.value === 'consumer' && <Connection id={event.node.id} consumer={event.node.data.consumer_details} />}
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
                        edge={(edge) => (
                            <Edge
                                {...edge}
                                className={edge?.data?.active === true ? (edge?.data?.is_producer ? 'edge produce-processing' : 'edge consume-processing') : 'edge'}
                            />
                        )}
                    />
                </div>
            )}
            {!isLoading && nodes?.length === 0 && (
                <div className="empty-connections-container">
                    <GraphPlaceholder alt="graphPlaceholder" onClick={() => createStationTrigger(true)} />
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
