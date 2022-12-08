// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useContext, useState } from 'react';

import { convertBytes, numberWithCommas, parsingDate } from '../../services/valueConvertor';
import PoisonMessage from './components/poisonMessage';
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
import { StringCodec, JSONCodec } from 'nats.ws';

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
            const data = await httpRequest('GET', `${ApiEndpoints.GET_POISON_MESSAGE_JOURNEY}?message_id=${messageId}`);
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
        const sub = state.socket?.subscribe(`$memphis_ws_pubs.poison_message_journey_data.${messageId}`);
        const jc = JSONCodec();
        const sc = StringCodec();
        if (sub) {
            (async () => {
                for await (const msg of sub) {
                    let data = jc.decode(msg.data);
                    arrangeData(data);
                }
            })();
        }

        setTimeout(() => {
            state.socket?.publish(`$memphis_ws_subs.poison_message_journey_data.${messageId}`, sc.encode('SUB'));
        }, 1000);
        return () => {
            sub?.unsubscribe();
        };
    }, [state.socket]);

    const returnBack = () => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };

    const arrangeData = (data) => {
        let poisonedCGs = [];
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
                width: 350,
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
                    width: 490,
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
                poisonedCGs.push(cg);
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
                headers: data.message?.headers,
                poisonedCGs: poisonedCGs
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
                        <img src={BackIcon} onClick={() => returnBack()} alt="backIcon" />
                        <p>
                            {stationName} / Poison message #{messageId.substring(0, 5)}
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
                                                <PoisonMessage
                                                    stationName={stationName}
                                                    messageId={messageId}
                                                    message={messageData.message}
                                                    headers={messageData.headers}
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
