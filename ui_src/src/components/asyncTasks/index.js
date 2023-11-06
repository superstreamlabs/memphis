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

import React, { useContext, useEffect, useState } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Divider, Popover } from 'antd';
import { parsingDate } from '../../services/valueConvertor';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { ReactComponent as AsyncIcon } from '../../assets/images/asyncIcon.svg';
import { ReactComponent as TaskIcon } from '../../assets/images/task.svg';
import { httpRequest } from '../../services/http';
import { ReactComponent as CollapseArrowIcon } from '../../assets/images/collapseArrow.svg';
import { Context } from '../../hooks/store';
import OverflowTip from '../tooltip/overflowtip';

const AsyncTasks = ({ height, overView }) => {
    const [state, dispatch] = useContext(Context);
    const [isOpen, setIsOpen] = useState(false);
    const [showMore, setShowMore] = useState(false);
    const [asyncTasks, setAsyncTasks] = useState([]);

    useEffect(() => {
        getAsyncTasks();
    }, []);

    useEffect(() => {
        let sub;
        let jc;
        let sc;

        const subscribeAndListen = async (subName, pubName, dataHandler) => {
            jc = JSONCodec();
            sc = StringCodec();

            try {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.${subName}`, sc.encode('SUB'));

                if (rawBrokerName) {
                    const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                    sub = state.socket?.subscribe(`$memphis_ws_pubs.${pubName}.${brokerName}`);
                }
            } catch (err) {
                console.error(`Error subscribing to ${subName} data:`, err);
                return;
            }

            setTimeout(async () => {
                if (sub) {
                    try {
                        for await (const msg of sub) {
                            let data = jc.decode(msg.data);
                            dataHandler(data);
                        }
                    } catch (err) {
                        console.error(`Error receiving ${subName} data updates:`, err);
                    }
                }
            }, 1000);
        };

        (async () => {
            try {
                await subscribeAndListen('get_async_tasks', 'get_async_tasks', (data) => {
                    setAsyncTasks(data);
                });
            } catch (err) {
                console.error('Error subscribing and listening to get_all_stations_data:', err);
            }
        })();

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from filters data:', err);
                }
            }
        };
    }, [state.socket]);

    const getAsyncTasks = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ASYNC_TASKS);
            data?.async_tasks?.length > 0 && setAsyncTasks(data.async_tasks);
        } catch (error) {}
    };

    const handleOpenChange = () => {
        setIsOpen(!isOpen);
    };

    const getItems = () => {
        return (
            <div>
                {asyncTasks.map((task, index) => {
                    return (
                        ((!showMore && index < 3) || showMore) && (
                            <div>
                                <div className="task-item" key={index}>
                                    <div>
                                        <TaskIcon alt="taskIcon" />
                                    </div>
                                    <div>
                                        <p className="task-title">
                                            {overView && `${task?.station_name} | `}
                                            {task?.name}
                                        </p>

                                        <OverflowTip width={'240px'} className="created" text={`Created by ${task?.created_by} at ${parsingDate(task?.created_at)}`}>
                                            <label className="created">
                                                Created by <b>{task?.created_by}</b> at {parsingDate(task?.created_at)}
                                            </label>
                                        </OverflowTip>
                                    </div>
                                </div>
                                <Divider />
                            </div>
                        )
                    );
                })}
            </div>
        );
    };
    const getContent = () => {
        return (
            <div>
                <div className="async-title">
                    <span>
                        <p>Async tasks</p>
                        <label className="async-number">{asyncTasks.length}</label>
                    </span>
                    <Divider />
                </div>
                <div className="tasks-container">{getItems()}</div>
                {asyncTasks.length > 3 && (
                    <div className="show-more-less-tasks" onClick={() => setShowMore(!showMore)}>
                        <label> {!showMore ? 'Show more' : 'Show less'}</label>
                    </div>
                )}
            </div>
        );
    };

    return (
        asyncTasks?.length > 0 && (
            <Popover placement="bottomLeft" content={getContent()} trigger="click" onOpenChange={handleOpenChange} open={isOpen}>
                <div className="async-btn-container">
                    <div className="async-btn">
                        <AsyncIcon alt="AsyncIcon" />
                        <div>
                            <label className="async-title">Async tasks </label>
                            <label className="async-number">{asyncTasks.length}</label>
                        </div>
                        <CollapseArrowIcon className={isOpen ? 'collapse-arrow open' : 'collapse-arrow'} alt="CollapseArrowIcon" />
                    </div>
                </div>
            </Popover>
        )
    );
};
export default AsyncTasks;
