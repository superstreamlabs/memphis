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

import { useState, useRef, useEffect, useCallback, useContext } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Virtuoso } from 'react-virtuoso';
import Lottie from 'lottie-react';

import { ReactComponent as AttachedPlaceholderIcon } from '../../../../assets/images/attachedPlaceholder.svg';
import animationData from '../../../../assets/lotties/MemphisGif.json';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Filter from '../../../../components/filter';
import { Context } from '../../../../hooks/store';
import { Sleep } from '../../../../utils/sleep';
import LogPayload from '../logPayload';
import LogContent from '../logContent';
import { LOGS_RETENTION_IN_DAYS } from '../../../../const/localStorageConsts';

let sub;

const LogsWrapper = () => {
    const [state, dispatch] = useContext(Context);
    const [displayedLog, setDisplayedLog] = useState({});
    const [selectedRow, setSelectedRow] = useState(null);
    const [visibleRange, setVisibleRange] = useState(0);
    const [logType, setLogType] = useState('external');
    const [logSource, setLogSource] = useState('empty');
    const [logs, setLogs] = useState(() => []);
    const [seqNum, setSeqNum] = useState(-1);
    const [stopLoad, setStopLoad] = useState(false);
    const [changed, setChanged] = useState(false);
    const [socketOn, setSocketOn] = useState(false);
    const [changeSelected, setChangeSelected] = useState(true);
    const [lastMgsSeq, setLastMgsSeq] = useState(-1);
    const [loader, setLoader] = useState(true);

    const stateRef = useRef([]);

    stateRef.current = [seqNum, visibleRange, socketOn, lastMgsSeq, changeSelected, logType, logSource];

    const getLogs = async (changed = false, seqNum = null) => {
        changed && setLoader(true);
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_SYS_LOGS}?log_type=${stateRef.current[5]}&log_source=${stateRef.current[6]}&start_index=${seqNum || stateRef.current[0]}`
            );
            if (data.logs && !changed) {
                if (stateRef.current[0] === -1) {
                    setLastMgsSeq(data.logs[0].message_seq);
                    setDisplayedLog(data.logs[0]);
                    setSelectedRow(data.logs[0].message_seq);
                }
                let message_seq = data.logs[data.logs.length - 1].message_seq;
                if (message_seq === stateRef.current[0]) {
                    setStopLoad(true);
                } else {
                    setSeqNum(message_seq);
                    setLogs((users) => [...users, ...data.logs]);
                }
            }
            if (changed && data.logs) {
                setLastMgsSeq(data.logs[0].message_seq);
                setDisplayedLog(data.logs[0]);
                setSelectedRow(data.logs[0].message_seq);
                let message_seq = data.logs[data.logs.length - 1].message_seq;
                setSeqNum(message_seq);
                setLogs(data.logs);
                setStopLoad(false);
                startListen();
            }
            if (changed && data.logs === null) {
                setLogs([]);
                setDisplayedLog({});
                await Sleep(1);
            }
            setLoader(false);
        } catch (error) {
            setLoader(false);
        }
    };

    const loadMore = useCallback(() => {
        return setTimeout(() => {
            getLogs();
        }, 200);
    }, []);

    useEffect(() => {
        const timeout = loadMore();
        return () => clearTimeout(timeout);
    }, []);

    useEffect(() => {
        if (stateRef.current[2]) {
            if (stateRef.current[1] !== 0) {
                stopListen();
            } else {
                startListen();
            }
        }
        return () => {};
    }, [stateRef.current[1]]);

    useEffect(() => {
        if (changed && sub) {
            stopListen();
            getLogs(changed, -1);
            setChanged(false);
        }
        return () => {};
    }, [sub, changed]);

    const startListen = async () => {
        const jc = JSONCodec();
        const sc = StringCodec();

        const subscribeToLogsWithouFilter = async () => {
            try {
                (async () => {
                    const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.syslogs_data`, sc.encode('SUB'));
                    if (rawBrokerName) {
                        const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                        sub = state.socket?.subscribe(`$memphis_ws_pubs.syslogs_data.${brokerName}`);
                        listenForUpdates();
                    }
                })();
            } catch (err) {
                console.error('Error subscribing to syslogs_data:', err);
            }
        };
        const subscribeToLogsWithFilter = async () => {
            try {
                (async () => {
                    let logFilter = `${logType}.${logSource}`;
                    if (logSource === '') {
                        logFilter = `${logType}`;
                    }
                    const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.syslogs_data.${logFilter}`, sc.encode('SUB'));
                    if (rawBrokerName) {
                        const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                        sub = state.socket?.subscribe(`$memphis_ws_pubs.syslogs_data.${logFilter}.${brokerName}`);
                        listenForUpdates();
                    }
                })();
            } catch (err) {
                console.error(`Error subscribing to syslogs_data_${logType}_${logSource}:`, err);
            }
        };
        if (logType === 'external' && logSource === '') {
            subscribeToLogsWithouFilter();
        } else {
            subscribeToLogsWithFilter();
        }

        const listenForUpdates = async () => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        let lastMgsSeqIndex = data.logs?.findIndex((log) => log.message_seq === stateRef.current[3]);
                        const uniqueItems = data.logs?.slice(0, lastMgsSeqIndex);
                        if (stateRef.current[4]) {
                            setSelectedRow(data?.logs[0]?.message_seq);
                            setDisplayedLog(data?.logs[0]);
                        }
                        setLastMgsSeq(data.logs[0].message_seq);
                        setLogs((users) => [...uniqueItems, ...users]);
                    }
                }
            } catch (err) {
                console.error(`Error receiving data updates for system logs:`, err);
            }
        };
    };

    const stopListen = () => {
        if (sub) {
            try {
                sub.unsubscribe();
            } catch (err) {
                console.error('Error unsubscribing from system logs data:', err);
            }
        }
    };

    useEffect(() => {
        if (state.socket) {
            setSocketOn(true);
            startListen();
        }
        return () => {
            stopListen();
        };
    }, [state.socket]);

    const selsectLog = (key) => {
        if (key === lastMgsSeq) {
            setChangeSelected(true);
        } else setChangeSelected(false);
        setSelectedRow(key);
        setDisplayedLog(logs.find((log) => log.message_seq === key));
    };

    const handleFilter = async (e) => {
        if (e[0] !== logType) {
            setLogType(e[0]);
            setChanged(true);
            setDisplayedLog({});
        }
        if (e[1] !== logSource) {
            setLogSource(e[1]);
            setChanged(true);
            setDisplayedLog({});
        }
    };

    return (
        <div className="logs-wrapper">
            <logs is="3xd">
                <list-header is="3xd">
                    <div className="header-title-wrapper">
                        <p className="header-title">Latest logs {logs?.length > 0 && `(${logs?.length})`}</p>
                        <Filter filterComponent="syslogs" height="34px" applyFilter={(e) => handleFilter(e)} />
                    </div>
                    <div className="header-subtitle">
                        <p>Logs will be retained for {localStorage.getItem(LOGS_RETENTION_IN_DAYS)} days</p>
                    </div>
                </list-header>
                {!loader && logs?.length > 0 && (
                    <Virtuoso
                        data={logs}
                        rangeChanged={(e) => setVisibleRange(e.startIndex)}
                        className="logsl"
                        endReached={!stopLoad ? loadMore : null}
                        overscan={100}
                        itemContent={(index, log) => (
                            <div className={index % 2 === 0 ? 'even' : 'odd'}>
                                <LogPayload selectedRow={selectedRow} value={log} onSelected={(e) => selsectLog(e)} />
                            </div>
                        )}
                        components={!stopLoad ? { Footer } : {}}
                    />
                )}
                {!loader && logs?.length === 0 && (
                    <div className="placeholder">
                        <AttachedPlaceholderIcon />
                        <p>No logs found</p>
                    </div>
                )}

                {loader && <div className="loader">{Footer()}</div>}
            </logs>
            <LogContent displayedLog={displayedLog} />
        </div>
    );
};

export default LogsWrapper;

const Footer = () => {
    return (
        <div
            className="logs-loader"
            style={{
                display: 'flex',
                justifyContent: 'center',
                height: '10vw'
            }}
        >
            <Lottie animationData={animationData} loop={true} />
        </div>
    );
};
