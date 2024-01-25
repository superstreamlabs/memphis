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

import React, { createContext, useContext, useEffect, useReducer, useState } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Popover } from 'antd';

import { filterType, labelType, CircleLetterColor } from 'const/globalConst';
import { ReactComponent as SearchIcon } from 'assets/images/searchIcon.svg';
import { ApiEndpoints } from 'const/apiEndpoints';
import { ReactComponent as FilterImgIcon } from 'assets/images/filter.svg';
import { httpRequest } from 'services/http';
import CustomCollapse from './customCollapse';
import { Context } from 'hooks/store';
import SearchInput from 'components/searchInput';
import Reducer from './hooks/reducer';
import Button from 'components/button';

const initialState = {
    isOpen: false,
    counter: 0,
    filterFields: []
};

const Filter = ({ filterComponent, height, applyFilter, hideElement }) => {
    const [state, dispatch] = useContext(Context);
    const [filterState, filterDispatch] = useReducer(Reducer, initialState);
    const [filterFields, setFilterFields] = useState([]);
    const [filterTerms, setFilterTerms] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    let sub;

    useEffect(() => {
        if (filterComponent === 'syslogs' && state?.logsFilter !== '') dispatch({ type: 'SET_LOG_FILTER', payload: ['', 'empty'] });
    }, [filterComponent]);

    useEffect(() => {
        if (filterState.isOpen && filterState.counter === 0) {
            getFilterDetails();
        }
    }, [filterState.isOpen]);

    useEffect(() => {
        filterFields.length > 0 && filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filterFields });
    }, [filterFields]);

    useEffect(() => {
        handleFilter();
    }, [searchInput, filterTerms, state?.stationList, state?.schemaList]);

    const getFilterDetails = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_FILTER_DETAILS}?route=${filterComponent}`);
            if (res) buildFilter(res);
        } catch (err) {
            return;
        }
    };

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

        switch (filterComponent) {
            case 'stations':
                (async () => {
                    try {
                        await subscribeAndListen('get_all_stations_data', 'get_all_stations_data', (data) => {
                            data?.sort((a, b) => new Date(b.station.created_at) - new Date(a.station.created_at));
                            dispatch({ type: 'SET_STATION_LIST', payload: data });
                        });
                    } catch (err) {
                        console.error('Error subscribing and listening to get_all_stations_data:', err);
                    }
                })();
                break;

            case 'schemaverse':
                (async () => {
                    try {
                        await subscribeAndListen('get_all_schema_data', 'get_all_schema_data', (data) => {
                            dispatch({ type: 'SET_SCHEMA_LIST', payload: data });
                        });
                    } catch (err) {
                        console.error('Error subscribing and listening to get_all_schema_data:', err);
                    }
                })();
                break;

            default:
                break;
        }

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from filters data:', err);
                }
            }
        };
    }, [filterComponent, state.socket]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const buildFilter = (rawFilterDetails) => {
        switch (filterComponent) {
            case 'stations':
                drawStationsFilter(rawFilterDetails);
                return;
            case 'syslogs':
                drawLogsFilter(rawFilterDetails);
                return;
            case 'schemaverse':
                drawSchemaFilter(rawFilterDetails);
                return;
            default:
                return;
        }
    };

    const drawStationsFilter = (rawFilterDetails) => {
        let filteredFields = [];
        if (rawFilterDetails?.tags?.length > 0) {
            const tagFilter = {
                name: 'tags',
                value: 'Tags',
                labelType: labelType.BADGE,
                filterType: filterType.CHECKBOX,
                fields: rawFilterDetails.tags
            };
            filteredFields.push(tagFilter);
        }

        const createdFilter = {
            name: 'created',
            value: 'Created By',
            labelType: labelType.CIRCLEDLETTER,
            filterType: filterType.CHECKBOX,
            fields: rawFilterDetails?.users?.map((user) => {
                return {
                    name: user,
                    color: CircleLetterColor[user[0]?.toUpperCase()],
                    checked: false
                };
            })
        };
        filteredFields.push(createdFilter);

        const storageTypeFilter = {
            name: 'storage',
            value: 'Storage Type',
            filterType: filterType.CHECKBOX,
            labelType: '',
            fields: rawFilterDetails?.storage?.map((s) => {
                return { name: s, value: s };
            })
        };
        filteredFields.push(storageTypeFilter);
        setFilterFields(filteredFields);
    };

    const drawSchemaFilter = (rawFilterDetails) => {
        let filteredFields = [];
        if (rawFilterDetails?.tags?.length > 0) {
            const tagFilter = {
                name: 'tags',
                value: 'Tags',
                labelType: labelType.BADGE,
                filterType: filterType.CHECKBOX,
                fields: rawFilterDetails.tags
            };
            filteredFields.push(tagFilter);
        }

        const createdFilter = {
            name: 'created',
            value: 'Created By',
            labelType: labelType.CIRCLEDLETTER,
            filterType: filterType.CHECKBOX,
            fields: rawFilterDetails?.users?.map((user) => {
                return {
                    name: user,
                    color: CircleLetterColor[user[0]?.toUpperCase()],
                    checked: false
                };
            })
        };
        filteredFields.push(createdFilter);

        const typeFilter = {
            name: 'type',
            value: 'Type',
            filterType: filterType.RADIOBUTTON,
            radioValue: -1,
            fields: rawFilterDetails?.type?.map((type) => {
                return { name: type };
            })
        };
        filteredFields.push(typeFilter);

        const usageFilter = {
            name: 'usage',
            value: 'Usage',
            filterType: filterType.RADIOBUTTON,
            radioValue: -1,
            fields: rawFilterDetails?.usage?.map((type) => {
                return { name: type };
            })
        };
        filteredFields.push(usageFilter);
        setFilterFields(filteredFields);
    };

    const drawLogsFilter = (rawFilterDetails) => {
        let filteredFields = [];
        const typeFilter = {
            name: 'type',
            value: 'Type',
            filterType: filterType.RADIOBUTTON,
            radioValue: -1,
            fields: rawFilterDetails?.type?.map((type) => {
                return { name: type };
            })
        };
        const sourceFilter = {
            name: 'source',
            value: 'Source',
            filterType: filterType.RADIOBUTTON,
            radioValue: -1,
            fields: rawFilterDetails?.source?.map((type) => {
                return { name: type };
            })
        };
        filteredFields.push(typeFilter, sourceFilter);
        setFilterFields(filteredFields);
    };

    const flipOpen = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: !filterState.isOpen });
    };

    const handleFilter = () => {
        let objTags = [];
        let objCreated = [];
        let objStorage = [];
        let objType = '';
        let objUsage = null;

        switch (filterComponent) {
            case 'stations':
                let stationData = state?.stationList;
                if (filterTerms?.find((o) => o?.name === 'tags')) {
                    objTags = filterTerms?.find((o) => o?.name === 'tags')?.fields?.map((element) => element?.toLowerCase());
                    stationData = stationData?.filter((item) =>
                        objTags?.length > 0 ? item.tags.some((tag) => objTags?.includes(tag?.name)) : !item.tags.some((tag) => objTags?.includes(tag?.name))
                    );
                }
                if (filterTerms?.find((o) => o?.name === 'created')) {
                    objCreated = filterTerms?.find((o) => o?.name === 'created')?.fields?.map((element) => element?.toLowerCase());
                    stationData = stationData?.filter((item) =>
                        objCreated?.length > 0 ? objCreated?.includes(item.station.created_by_username) : !objCreated?.includes(item.station.created_by_username)
                    );
                }
                if (filterTerms?.find((o) => o?.name === 'storage')) {
                    objStorage = filterTerms?.find((o) => o?.name === 'storage')?.fields?.map((element) => element?.toLowerCase());
                    stationData = stationData.filter((item) =>
                        objStorage?.length > 0 ? objStorage?.includes(item.station.storage_type) : !objStorage?.includes(item.station.storage_type)
                    );
                }
                if (searchInput !== '' && searchInput?.length >= 2) {
                    stationData = stationData.filter((station) => station.station?.name?.includes(searchInput));
                }
                dispatch({ type: 'SET_STATION_FILTERED_LIST', payload: stationData });
                return;
            case 'schemaverse':
                let data = state?.schemaList;
                if (filterTerms?.find((o) => o?.name === 'tags')) {
                    objTags = filterTerms?.find((o) => o?.name === 'tags')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) =>
                        objTags?.length > 0 ? item.tags.some((tag) => objTags?.includes(tag?.name)) : !item.tags.some((tag) => objTags?.includes(tag?.name))
                    );
                }
                if (filterTerms?.find((o) => o?.name === 'created')) {
                    objCreated = filterTerms?.find((o) => o?.name === 'created')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) =>
                        objCreated?.length > 0 ? objCreated?.includes(item.created_by_username) : !objCreated?.includes(item.created_by_username)
                    );
                }
                if (filterTerms?.find((o) => o?.name === 'type')) {
                    objType = filterTerms?.find((o) => o?.name === 'type')?.fields[0];
                    data = data?.filter((item) => objType !== '' && item.type === objType);
                }
                if (filterTerms?.find((o) => o?.name === 'usage')) {
                    objUsage = filterTerms?.find((o) => o?.name === 'usage')?.fields[0] === 'used';
                    data = data.filter((item) => item.used === objUsage);
                }
                if (searchInput !== '' && searchInput?.length >= 2) {
                    data = data.filter((schema) => schema?.name?.includes(searchInput));
                }
                dispatch({ type: 'SET_SCHEMA_FILTERED_LIST', payload: data });
                return;
            default:
                return;
        }
    };

    const handleApply = () => {
        if (filterComponent === 'syslogs') {
            const selectedTypeField = filterState?.filterFields[0]?.radioValue;
            const selectedSourceField = filterState?.filterFields[1]?.radioValue;
            if (selectedTypeField !== -1 && selectedSourceField !== -1) {
                dispatch({
                    type: 'SET_LOG_FILTER',
                    payload: [filterState?.filterFields[0]?.fields[selectedTypeField]?.name, filterState?.filterFields[1]?.fields[selectedSourceField]?.name]
                });
                applyFilter([filterState?.filterFields[0]?.fields[selectedTypeField]?.name, filterState?.filterFields[1]?.fields[selectedSourceField]?.name]);
                setFilterTerms([filterState?.filterFields[0]?.fields[selectedTypeField]?.name, filterState?.filterFields[1]?.fields[selectedSourceField]?.name]);
            } else if (selectedTypeField !== -1 && selectedSourceField === -1) {
                dispatch({ type: 'SET_LOG_FILTER', payload: [filterState?.filterFields[0]?.fields[selectedTypeField]?.name, 'empty'] });
                applyFilter([filterState?.filterFields[0]?.fields[selectedTypeField]?.name, 'empty']);
                setFilterTerms([filterState?.filterFields[0]?.fields[selectedTypeField]?.name, 'empty']);
            } else if (selectedTypeField === -1 && selectedSourceField !== -1) {
                dispatch({ type: 'SET_LOG_FILTER', payload: ['external', filterState?.filterFields[1]?.fields[selectedSourceField]?.name] });
                applyFilter(['external', filterState?.filterFields[1]?.fields[selectedSourceField]?.name]);
                setFilterTerms(['external', filterState?.filterFields[1]?.fields[selectedSourceField]?.name]);
            } else {
                dispatch({ type: 'SET_LOG_FILTER', payload: ['external', 'empty'] });
                applyFilter(['external', 'empty']);
                setFilterTerms(['external', 'empty']);
            }
        } else {
            let filterTerms = [];
            filterState?.filterFields.forEach((element) => {
                let term = {
                    name: element?.name,
                    fields: []
                };
                if (element.filterType === filterType.CHECKBOX) {
                    element.fields.forEach((field) => {
                        if (field.checked) {
                            let t = term.fields;
                            t.push(field?.name);
                            term.fields = t;
                        }
                    });
                } else if (element.filterType === filterType.RADIOBUTTON && element.radioValue !== -1) {
                    let t = [];
                    t.push(element.fields[element.radioValue]?.name);
                    term.fields = t;
                } else {
                    element.fields.forEach((field) => {
                        if (field?.value !== undefined && field?.value !== '') {
                            let t = term.fields;
                            let d = {};
                            d[field?.name] = field.value;
                            t.push(d);
                            term.fields = t;
                        }
                    });
                }
                if (term.fields.length > 0) filterTerms.push(term);
            });
            setFilterTerms(filterTerms);
        }
        flipOpen();
    };

    const handleClear = () => {
        filterDispatch({ type: 'SET_COUNTER', payload: 0 });
        let filter = filterFields;
        filter.map((filterGroup) => {
            switch (filterGroup.filterType) {
                case filterType.CHECKBOX:
                    filterGroup.fields.map((field) => (field.checked = false));
                case filterType.DATE:
                    filterGroup.fields.map((field) => (field.value = ''));
                case filterType.RADIOBUTTON:
                    filterGroup.radioValue = -1;
            }
        });
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filter });
        setFilterTerms([]);
        if (filterComponent === 'syslogs') {
            dispatch({ type: 'SET_LOG_FILTER', payload: ['external', 'empty'] });
            applyFilter(['external', 'empty']);
        }
    };

    const handleCancel = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: false });
    };

    const handleOpenChange = () => {
        flipOpen();
    };

    const content = <CustomCollapse header="Details" data={filterState?.filterFields} cancel={handleCancel} apply={handleApply} clear={handleClear} />;

    return (
        <FilterStoreContext.Provider value={[filterState, filterDispatch]}>
            {filterComponent !== 'syslogs' && (
                (hideElement !== 'search' &&
                <SearchInput
                    placeholder="Search"
                    colorType="navy"
                    backgroundColorType="gray-dark"
                    width={ hideElement === 'search' ? "288px" : "100%" }
                    height="34px"
                    borderColorType="none"
                    boxShadowsType="none"
                    borderRadiusType="circle"
                    iconComponent={<SearchIcon alt="searchIcon" />}
                    onChange={handleSearch}
                    value={searchInput}
                />)
            )}
            {(hideElement !== 'filter' &&
            <Popover placement="bottomLeft" content={content} trigger="click" onOpenChange={handleOpenChange} open={filterState.isOpen}>
                <Button
                    className="modal-btn"
                    width="110px"
                    height={height}
                    placeholder={
                        <div className="filter-container">
                            <FilterImgIcon alt="filter" width={25} />
                            <label className="filter-title">Filters</label>
                            {filterTerms?.length > 0 && filterState?.counter > 0 && <div className="filter-counter">{filterState?.counter}</div>}
                        </div>
                    }
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="12px"
                    fontWeight="bold"
                    boxShadowStyle="float"
                    onClick={() => {}}
                />
            </Popover>)}
        </FilterStoreContext.Provider>
    );
};
export const FilterStoreContext = createContext({});
export default Filter;
