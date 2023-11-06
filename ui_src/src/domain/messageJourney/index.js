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

import React, { useEffect, useContext, useState } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { useHistory } from 'react-router-dom';

import { convertBytes, extractValueFromURL, parsingDate } from '../../services/valueConvertor';
import PoisonMessage from './components/poisonMessage';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { ReactComponent as BackIcon } from '../../assets/images/backIcon.svg';
import ConsumerGroup from './components/consumerGroup';
import { Canvas, Node, Edge, Label } from 'reaflow';
import { httpRequest } from '../../services/http';
import Producer from './components/producer';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import pathDomains from '../../router';
import { showMessages } from '../../services/genericServices';

const MessageJourney = () => {
    const [state, dispatch] = useContext(Context);
    const url = window.location.href;
    const messageId = extractValueFromURL(url, 'id');
    const stationName = extractValueFromURL(url, 'name');
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
        let sub;
        const jc = JSONCodec();
        const sc = StringCodec();

        const subscribeAndListen = async (subName, pubName, messageId) => {
            try {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.${subName}.${messageId}`, sc.encode('SUB'));

                if (rawBrokerName) {
                    const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                    sub = state.socket?.subscribe(`$memphis_ws_pubs.${pubName}.${messageId}.${brokerName}`);
                    listenForUpdates(subName, messageId);
                }
            } catch (err) {
                console.error(`Error subscribing to ${subName} data for messageId ${messageId}:`, err);
                return;
            }
        };
        const listenForUpdates = async (subName, messageId) => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        arrangeData(data);
                    }
                }
            } catch (err) {
                console.error(`Error receiving ${subName} data updates for messageId ${messageId}:`, err);
            }
        };

        subscribeAndListen('poison_message_journey_data', 'poison_message_journey_data', messageId);

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from message journey data:', err);
                }
            }
        };
    }, [messageId, state.socket]);

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
            if (!data?.poisoned_cgs || data?.poisoned_cgs.length === 0) {
                showMessages('success', 'Unacknowledged message has been acked by all of its failed CGs');
                returnBack();
            }
            data?.poisoned_cgs?.map((row, index) => {
                let cg = {
                    name: row.cg_name,
                    is_active: row.is_active,
                    is_deleted: row.is_deleted,
                    cgMembers: row.cg_members,
                    details: [
                        {
                            name: 'Unacknowledged messages',
                            value: row?.total_poison_messages?.toLocaleString()
                        },
                        {
                            name: 'Unprocessed messages',
                            value: row?.unprocessed_messages?.toLocaleString()
                        },
                        {
                            name: 'In process message',
                            value: row?.in_process_messages?.toLocaleString()
                        },
                        {
                            name: 'Max ack time',
                            value: `${row?.max_ack_time_ms?.toLocaleString()}ms`
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
                                name: 'Unacknowledged messages',
                                value: row.total_poison_messages?.toLocaleString()
                            },
                            {
                                name: 'Unprocessed messages',
                                value: row.unprocessed_messages?.toLocaleString()
                            },
                            {
                                name: 'In process message',
                                value: row.in_process_messages?.toLocaleString()
                            },
                            {
                                name: 'Max ack time',
                                value: `${row.max_ack_time_ms?.toLocaleString()}ms`
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
                _id: data._id ?? null,
                message_seq: data.message_seq,
                details: [
                    {
                        name: 'Message size',
                        value: convertBytes(data.message?.size)
                    },
                    {
                        name: 'Time sent',
                        value: parsingDate(data.message?.time_sent, true)
                    }
                ],
                producer: {
                    is_active: data.producer?.is_active,
                    is_deleted: data.producer?.is_deleted,
                    name: data.producer?.name
                },
                message: data.message?.data,
                headers: data.message?.headers,
                poisonedCGs: poisonedCGs,
                schemaType: data?.schema_type
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
                        <BackIcon alt="backIcon" onClick={() => returnBack()} />
                        <p>
                            Message seq: <span>{messageData?.message_seq}</span>
                        </p>
                    </div>
                    <div className="canvas-wrapper">
                        <Canvas
                            className="canvas"
                            readonly={true}
                            direction="RIGHT"
                            nodes={nodes}
                            edges={edges}
                            fit={true}
                            height={'100%'}
                            maxHeight={nodes?.length < 5 ? 700 : nodes?.length * 170}
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
                                                    schemaType={messageData.schemaType}
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
