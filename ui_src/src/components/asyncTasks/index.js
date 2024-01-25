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
import { Divider, Popover, Badge } from 'antd';
import { parsingDate } from 'services/valueConvertor';
import { ApiEndpoints } from 'const/apiEndpoints';
import { ReactComponent as BgTasksIcon } from 'assets/images/bgTasksIcon.svg';
import { ReactComponent as TaskIcon } from 'assets/images/task.svg';
import { httpRequest } from 'services/http';
import { Context } from 'hooks/store';
import OverflowTip from 'components/tooltip/overflowtip';

const AsyncTasks = ({ overView, children }) => {
    const [state, dispatch] = useContext(Context);
    const [isOpen, setIsOpen] = useState(false);
    const [showMore, setShowMore] = useState(false);
    const [asyncTasks, setAsyncTasks] = useState([]);
    const [runningTasks, setRunningTasks] = useState([]);

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

    useEffect(() => {
        const running = asyncTasks?.filter((task) => task?.status === 'running');
        setRunningTasks(running);
        dispatch({ type: 'SET_BACKGROUND_TASKS_COUNT', payload: running?.length || 0 });
    }, [asyncTasks]);

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
                                    <span className="task-status">
                                        <Badge status={task?.status === 'completed' ? 'success' : task?.status === 'failed' ? 'error' : 'processing'} />
                                    </span>
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
            asyncTasks?.length > 0 && (
                <div>
                    <div className="async-title">
                        <span>
                            <p>Background tasks</p>
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
            )
        );
    };

    return (
        <Popover placement={overView ? 'bottomRight' : 'bottomLeft'} content={getContent()} trigger="click" onOpenChange={handleOpenChange} open={isOpen}>
            {overView ? (
                <span className={asyncTasks?.length > 0 ? 'overview-tasks' : undefined}>{children}</span>
            ) : (
                runningTasks?.length > 0 && (
                    <div className="async-btn-container">
                        <BgTasksIcon alt="AsyncIcon" />
                    </div>
                )
            )}
        </Popover>
    );
};
export default AsyncTasks;
