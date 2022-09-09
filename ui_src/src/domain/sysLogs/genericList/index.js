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

import React, { Fragment, useContext, useEffect, useRef, useState } from 'react';

import OverflowTip from '../../../components/tooltip/overflowtip';
import { parsingDate } from '../../../services/valueConvertor';
import SearchInput from '../../../components/searchInput';
import { SearchOutlined } from '@ant-design/icons';
import Loader from '../../../components/loader';
import LogBadge from '../../../components/logBadge';
import SelectComponent from '../../../components/select';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { Context } from '../../../hooks/store';

const GenericList = ({ columns }) => {
    const types = ['all', 'info', 'warn', 'error'];

    const [logsState, setLogState] = useState({
        logsData: [],
        dataLength: 0,
        logFilter: 'all',
        searchInput: ''
    });
    const [state, dispatch] = useContext(Context);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedRowIndex, setSelectedRowIndex] = useState(0);
    const [logFilter, setLogFilter] = useState('all');
    const [searchInput, setSearchInput] = useState('');

    const stateRef = useRef([]);
    stateRef.current = [logsState.logsData, logFilter, searchInput];

    useEffect(() => {
        getSystemLogs(8);
    }, []);

    const getSystemLogs = async (hours) => {
        setIsLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_SYS_LOGS, {}, {}, { hours: hours });
            if (data?.logs) {
                let SortData = data.logs?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date)).map((log) => ({ ...log, show: true }));
                setLogState({ ...logsState, logsData: SortData, dataLength: SortData?.length, gotData: true });
            }
        } catch (error) {}
        setIsLoading(false);
    };

    const handleFilter = (data, comingLogs) => {
        if (data?.length > 0) {
            let counter = 0;
            setIsLoading(true);
            let result;
            if (stateRef.current[1] === 'all') {
                if (stateRef.current[2]?.length > 1) {
                    result = data?.map((log) => {
                        if (log?.log.toLowerCase().includes(stateRef.current[2])) {
                            counter++;
                            return { ...log, show: true };
                        } else {
                            return { ...log, show: false };
                        }
                    });
                } else {
                    let newData = data?.map((log) => {
                        counter++;
                        return { ...log, show: true };
                    });
                    result = newData;
                }
            } else if (stateRef.current[2]?.length > 1) {
                result = data?.map((log) => {
                    if (log.type.toLowerCase() === stateRef.current[1] && log?.log.toLowerCase().includes(stateRef.current[2])) {
                        counter++;
                        return { ...log, show: true };
                    } else {
                        return { ...log, show: false };
                    }
                });
            } else {
                result = data?.map((log) => {
                    if (log.type.toLowerCase() === stateRef.current[1]) {
                        counter++;
                        return { ...log, show: true };
                    } else {
                        return { ...log, show: false };
                    }
                });
            }
            if (comingLogs) {
                const newList = result.concat(stateRef.current[0]);
                setLogState({ ...logsState, logsData: newList, dataLength: newList?.length });
                setIsLoading(false);
            } else {
                setLogState({ ...logsState, logsData: result, dataLength: counter });
                setIsLoading(false);
            }
        }
    };

    useEffect(() => {
        state.socket?.on('system_logs_data', (data) => {
            if (data) {
                let sortData = data?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
                handleFilter(sortData, true);
            }
        });
        setTimeout(() => {
            state.socket?.emit('register_system_logs_data');
        }, 2000);

        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    useEffect(() => {
        handleFilter(logsState.logsData, false);
    }, [logFilter]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const onPressEnter = (e) => {
        e.preventDefault();
        handleFilter(logsState.logsData, false);
    };
    const onSelectedRow = (rowIndex) => {
        setSelectedRowIndex(rowIndex);
    };

    const updateLogFilter = (value) => {
        setLogFilter(value);
    };

    return (
        <Fragment>
            {isLoading && (
                <div>
                    <Loader />
                </div>
            )}
            {!isLoading && (
                <div className="logs-list-container">
                    <div className="add-search-logs">
                        <SearchInput
                            value={searchInput}
                            placeholder="Search logs"
                            colorType="navy"
                            backgroundColorType="none"
                            width="20vw"
                            height="28px"
                            borderRadiusType="circle"
                            borderColorType="gray"
                            boxShadowsType="gray"
                            iconComponent={<SearchOutlined />}
                            onChange={handleSearch}
                            onPressEnter={onPressEnter}
                        />
                    </div>
                    <div className="logs-wrapper">
                        <div className="logs-number">Showing {logsState.dataLength} logs in the last 8 hours</div>
                        <div className="logs-dropdown">
                            <SelectComponent
                                value={logFilter}
                                colorType="navy"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="circle"
                                width="100px"
                                height="28px"
                                options={types}
                                dropdownClassName="select-options"
                                onChange={(e) => updateLogFilter(e)}
                            />
                        </div>
                        <div className="logs-list-wrapper">
                            <div className="list">
                                <div className="coulmns-table">
                                    {columns?.map((column, index) => {
                                        return (
                                            <span key={index} style={{ width: column.width }}>
                                                {column.title}
                                            </span>
                                        );
                                    })}
                                </div>
                                <div className="rows-wrapper">
                                    {logsState.logsData?.length > 0 &&
                                        logsState.logsData?.map((row, index) => {
                                            if (row.show === true) {
                                                return (
                                                    <div
                                                        className={selectedRowIndex === index ? 'log-row selected' : 'log-row'}
                                                        key={index}
                                                        onClick={() => onSelectedRow(index)}
                                                    >
                                                        <OverflowTip text={row?.component} width={'100px'}>
                                                            {row?.component}
                                                        </OverflowTip>
                                                        <LogBadge type={row?.type}></LogBadge>
                                                        <OverflowTip text={parsingDate(row?.creation_date)} width={'200px'}>
                                                            {parsingDate(row?.creation_date)}
                                                        </OverflowTip>
                                                        <div className="log-field">{row?.log}</div>
                                                    </div>
                                                );
                                            }
                                        })}
                                </div>
                            </div>
                            <div className="row-data">
                                <p className="row-content">
                                    {logsState.logsData?.length > 0 && logsState.logsData[selectedRowIndex]?.show && logsState.logsData[selectedRowIndex]?.log}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </Fragment>
    );
};

export default GenericList;
