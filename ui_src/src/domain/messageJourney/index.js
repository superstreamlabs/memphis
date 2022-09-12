// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useEffect, useContext, useState } from 'react';

import { convertBytes, numberWithCommas, parsingDate } from '../../services/valueConvertor';
import PoisionMessage from './components/poisionMessage';
import { ApiEndpoints } from '../../const/apiEndpoints';
import BackIcon from '../../assets/images/backIcon.svg';
import ConsumerGroup from './components/consumerGroup';
import { Canvas, Node, Edge, Label } from 'reaflow';
import { httpRequest } from '../../services/http';
import { useHistory } from 'react-router-dom';
import Producer from './components/producer';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import pathDomains from '../../router';

const MessageJourney = () => {
    const [state, dispatch] = useContext(Context);
    const url = window.location.href;
    const messageId = url.split('stations/')[1].split('/')[1];
    const stationName = url.split('stations/')[1].split('/')[0];
    const [isLoading, setisLoading] = useState(false);
    const [processing, setProcessing] = useState(false);
    const [messageData, setMessageData] = useState({});
    const [nodes, setNodes] = useState();
    const [edges, setEdges] = useState();

    const history = useHistory();

    const getPosionMessageDetails = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_POISION_MESSAGE_JOURNEY}?message_id=${messageId}`);
            arrangeData(data);
        } catch (error) {
            setisLoading(false);
            if (error.status === 404 || error.status === 666) {
                returnBack();
            }
        }
    };

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'stations' });
        getPosionMessageDetails();
    }, []);

    useEffect(() => {
        state.socket?.on(`poison_message_journey_data_${messageId}`, (data) => {
            arrangeData(data);
        });
        setTimeout(() => {
            state.socket?.emit('register_poison_message_journey_data', messageId);
        }, 1000);
        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const returnBack = () => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };
    const arrangeData = (data) => {
        let poisionedCGs = [];
        let nodesList = [
            {
                id: 1,
                text: 'Node 1',
                width: 300,
                height: 170,
                data: {
                    value: 'producer'
                }
            },
            {
                id: 2,
                text: 'Node 2',
                width: 300,
                height: 600,
                ports: [
                    {
                        id: 'station',
                        side: 'EAST',
                        width: 10,
                        height: 10,
                        hidden: true
                    }
                ],
                data: {
                    value: 'station'
                }
            }
        ];
        let edgesList = [
            {
                id: 1,
                from: 1,
                to: 2,
                fromPort: 1,
                toPort: 2,
                selectionDisabled: true,
                data: {
                    value: 'producer'
                }
            }
        ];
        if (data) {
            data.poisoned_cgs.map((row, index) => {
                let cg = {
                    name: row.cg_name,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                    cgMembers: row.cg_members,
                    details: [
                        {
                            name: 'Poison messages',
                            value: numberWithCommas(row?.total_poison_messages)
                        },
                        {
                            name: 'Unprocessed messages',
                            value: numberWithCommas(row?.unprocessed_messages)
                        },
                        {
                            name: 'In process message',
                            value: numberWithCommas(row?.in_process_messages)
                        },
                        {
                            name: 'Max ack time',
                            value: `${numberWithCommas(row?.max_ack_time_ms)}ms`
                        },
                        {
                            name: 'Max message deliveries',
                            value: row?.max_msg_deliveries
                        }
                    ]
                };
                let node = {
                    id: index + 3,
                    text: row.cg_name,
                    width: 450,
                    height: 260,
                    data: {
                        value: 'consumer',
                        cgData: [
                            {
                                name: 'Poison messages',
                                value: numberWithCommas(row.total_poison_messages)
                            },
                            {
                                name: 'Unprocessed messages',
                                value: numberWithCommas(row.unprocessed_messages)
                            },
                            {
                                name: 'In process message',
                                value: numberWithCommas(row.in_process_messages)
                            },
                            {
                                name: 'Max ack time',
                                value: `${numberWithCommas(row.max_ack_time_ms)}ms`
                            },
                            {
                                name: 'Max message deliveries',
                                value: row.max_msg_deliveries
                            }
                        ],
                        cgMembers: row.cg_members
                    }
                };
                let edge = {
                    id: index + 2,
                    from: 2,
                    to: index + 3,
                    fromPort: 'station',
                    toPort: index + 3,
                    selectionDisabled: true,
                    data: {
                        value: 'consumer'
                    }
                };
                nodesList.push(node);
                edgesList.push(edge);
                poisionedCGs.push(cg);
            });

            let messageDetails = {
                id: data._id ?? null,
                messageSeq: data.message_seq,
                details: [
                    {
                        name: 'Message size',
                        value: convertBytes(data.message?.size)
                    },
                    {
                        name: 'Time sent',
                        value: parsingDate(data.message?.time_sent)
                    }
                ],
                producer: {
                    is_active: data.producer?.is_active,
                    is_deleted: data.producer?.is_deleted,
                    details: [
                        {
                            name: 'Name',
                            value: data.producer?.name
                        },
                        {
                            name: 'User',
                            value: data.producer?.created_by_user
                        },
                        {
                            name: 'IP',
                            value: data.producer?.client_address
                        }
                    ]
                },
                message: data.message?.data,
                poisionedCGs: poisionedCGs
            };
            setMessageData(messageDetails);
            setEdges(edgesList);
            setNodes(nodesList);
            setTimeout(() => {
                setisLoading(false);
            }, 1000);
        }
    };

    return (
        <>
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!isLoading && (
                <div className="message-journey-container">
                    <div className="bread-crumbs">
                        <img src={BackIcon} onClick={() => returnBack()} />
                        <p>
                            {stationName} / Poision message #{messageId.substring(0, 5)}
                        </p>
                    </div>

                    <div className="canvas-wrapper">
                        <Canvas
                            className="canvas"
                            readonly={true}
                            direction="RIGHT"
                            // defaultPosition={null}
                            nodes={nodes}
                            edges={edges}
                            node={
                                <Node style={{ stroke: 'transparent', fill: 'transparent', strokeWidth: 1 }} label={<Label style={{ display: 'none' }} />}>
                                    {(event) => (
                                        <foreignObject height={event.height} width={event.width} x={0} y={0} className="node-wrapper">
                                            {event.node.data.value === 'producer' && <Producer data={messageData.producer} />}
                                            {event.node.data.value === 'station' && (
                                                <PoisionMessage
                                                    stationName={stationName}
                                                    messageId={messageId}
                                                    message={messageData.message}
                                                    details={messageData.details}
                                                    processing={(status) => setProcessing(status)}
                                                    returnBack={() => returnBack()}
                                                />
                                            )}
                                            {event.node.data.value === 'consumer' && (
                                                <ConsumerGroup header={event.node.text} details={event.node.data.cgData} cgMembers={event.node.data.cgMembers} />
                                            )}
                                        </foreignObject>
                                    )}
                                </Node>
                            }
                            arrow={null}
                            edge={(edge) => (
                                <Edge
                                    {...edge}
                                    className={edge.data.value === 'producer' ? 'edge producer' : processing ? 'edge consumer processing' : 'edge consumer'}
                                />
                            )}
                        />
                    </div>
                </div>
            )}
        </>
    );
};
export default MessageJourney;
