// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useContext, useState, useRef } from 'react';
import { Canvas, Node, Edge, Label, MarkerArrow, Port } from 'reaflow';
import ConnectionGraphView from './components/connection';
import StationGraphView from './components/station';

function GraphView() {
    const [nodes, setNodes] = useState([
        {
            id: '1',
            text: 'Connection 10,9,0',
            data: {
                value: 'connection',
                port: '10.9.9',
                producers: ['01', '02'],
                consumers: ['03', '04'],
                createdBy: 'root'
            }
        },
        {
            id: '11',
            text: 'Connection 10,9,0',
            data: {
                value: 'connection',
                port: '10.9.9',
                producers: ['01', '02'],
                consumers: ['03', '04'],
                createdBy: 'root'
            }
        },
        {
            id: '2',
            text: 'Station 1',
            data: {
                value: 'station',
                name: 'station 01',
                totalMessages: 100,
                totalPoison: 10,
                createdBy: 'root'
            },
            ports: [
                {
                    id: 'station-east',
                    side: 'EAST',
                    width: 10,
                    height: 10,
                    hidden: true
                },
                {
                    id: 'station-west',
                    side: 'WEST',
                    width: 10,
                    height: 10,
                    hidden: true
                }
            ]
        },
        {
            id: '3',
            text: 'Connection 10,9,0',
            data: {
                value: 'connection',
                port: '10.9.9',
                producers: ['01', '02'],
                consumers: ['03', '04'],
                createdBy: 'root'
            }
        }
    ]);
    const [edges, setEdges] = useState([
        {
            id: '1-2',
            from: '1',
            to: '2',
            toPort: 'station-west',
            data: {
                active: true
            }
        },
        {
            id: '11-2',
            from: '11',
            to: '2',
            toPort: 'station-west',
            data: {
                active: true
            }
        },
        {
            id: '2-3',
            from: '2',
            to: '3',
            fromPort: 'station-east',
            data: {
                active: false
            }
        }
    ]);
    return (
        <div className="graph-view-container">
            <Canvas
                direction="RIGHT"
                fit={true}
                pannable={false}
                nodes={nodes}
                height={'100%'}
                edges={edges}
                node={
                    <Node style={{ stroke: 'transparent', fill: 'transparent', strokeWidth: 1 }} label={<Label style={{ display: 'none' }} />}>
                        {(event) => (
                            <foreignObject width={300} height={300} className="node-wrapper">
                                {event.node.data.value === 'connection' && <ConnectionGraphView data={event.node.data} />}
                                {event.node.data.value === 'station' && <StationGraphView data={event.node.data} />}
                            </foreignObject>
                        )}
                    </Node>
                }
                arrow={null}
                edge={(edge) => <Edge {...edge} className={edge.data.active ? 'edge active' : 'edge'} />}
            />
        </div>
    );
}

export default GraphView;
