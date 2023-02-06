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

import React, { createContext, useContext, useEffect, useReducer, useState } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Popover } from 'antd';

import { filterType, labelType, CircleLetterColor } from '../../const/globalConst';
import searchIcon from '../../assets/images/searchIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import filterImg from '../../assets/images/filter.svg';
import { httpRequest } from '../../services/http';
import CustomCollapse from './customCollapse';
import { Context } from '../../hooks/store';
import SearchInput from '../searchInput';
import Reducer from './hooks/reducer';
import Button from '../button';

const initialState = {
    isOpen: false,
    counter: 0,
    filterFields: []
};

const Filter = ({ filterComponent, height, applyFilter }) => {
    const [state, dispatch] = useContext(Context);
    const [filterState, filterDispatch] = useReducer(Reducer, initialState);
    const [filterFields, setFilterFields] = useState([]);
    const [filterTerms, setFilterTerms] = useState([]);
    const [searchInput, setSearchInput] = useState('');

    useEffect(() => {
        if (filterComponent === 'syslogs' && state?.logsFilter !== '') dispatch({ type: 'SET_LOG_FILTER', payload: '' });
    }, [filterComponent]);

    useEffect(() => {
        if (filterState.isOpen && filterState.counter === 0) {
            getFilterFields();
        }
    }, [filterState.isOpen]);

    useEffect(() => {
        //Initial the state with filterFields
        filterFields.length > 0 && filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filterFields });
    }, [filterFields]);

    useEffect(() => {
        handleFilter();
    }, [searchInput, filterTerms, state?.domainList]);

    const getFilterFields = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_FILTER_DETAILS}?route=${filterComponent}`);
            if (res) buildEmptyFilter(res);
        } catch (err) {
            return;
        }
    };

    useEffect(() => {
        let sub;
        let jc;
        let sc;
        switch (filterComponent) {
            case 'stations':
                jc = JSONCodec();
                sc = StringCodec();
                try {
                    (async () => {
                        const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.get_all_stations_data`, sc.encode('SUB'));
                        const brokerName = JSON.parse(sc.decode(rawBrokerName._rdata))['name'];
                        sub = state.socket?.subscribe(`$memphis_ws_pubs.get_all_stations_data.${brokerName}`);
                    })();
                } catch (err) {
                    return;
                }
                setTimeout(async () => {
                    if (sub) {
                        (async () => {
                            for await (const msg of sub) {
                                let data = jc.decode(msg.data);
                                data?.sort((a, b) => new Date(b.station.creation_date) - new Date(a.station.creation_date));
                                dispatch({ type: 'SET_DOMAIN_LIST', payload: data });
                            }
                        })();
                    }
                }, 1000);

                return () => {
                    sub?.unsubscribe();
                };
            case 'schemaverse':
                jc = JSONCodec();
                sc = StringCodec();
                try {
                    (async () => {
                        const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.get_all_schema_data`, sc.encode('SUB'));
                        const brokerName = JSON.parse(sc.decode(rawBrokerName._rdata))['name'];
                        sub = state.socket?.subscribe(`$memphis_ws_pubs.get_all_schema_data.${brokerName}`);
                    })();
                } catch (err) {
                    return;
                }
                setTimeout(async () => {
                    if (sub) {
                        (async () => {
                            for await (const msg of sub) {
                                let data = jc.decode(msg.data);
                                dispatch({ type: 'SET_DOMAIN_LIST', payload: data });
                            }
                        })();
                    }
                }, 1000);

                return () => {
                    sub?.unsubscribe();
                };
        }
    }, [state.socket]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const buildEmptyFilter = (rawFilterDetails) => {
        switch (filterComponent) {
            case 'stations':
                buildEmptyStationsFilter(rawFilterDetails);
                return;
            case 'syslogs':
                buildEmptyLogsFilter(rawFilterDetails);
                return;
            case 'schemaverse':
                buildEmptySchemaFilter(rawFilterDetails);
                return;
            default:
                return;
        }
    };

    const buildEmptyStationsFilter = (rawFilterDetails) => {
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
                return { name: user, color: CircleLetterColor[user[0]?.toUpperCase()], checked: false };
            })
        };
        filteredFields.push(createdFilter);

        const storageTypeFilter = {
            name: 'storage',
            value: 'Storage Type',
            filterType: filterType.CHECKBOX,
            labelType: '',
            fields: rawFilterDetails?.storage?.map((s) => {
                return { name: s, value: s, checked: false };
            })
        };
        filteredFields.push(storageTypeFilter);
        setFilterFields(filteredFields);
    };

    const buildEmptySchemaFilter = (rawFilterDetails) => {
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

    const buildEmptyLogsFilter = (rawFilterDetails) => {
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
        filteredFields.push(typeFilter);
        setFilterFields(filteredFields);
    };

    const flipOpen = () => {
        filterDispatch({ type: 'SET_IS_OPEN', payload: !filterState.isOpen });
    };

    const createFieldsToFilterObj = () => {
        let filterTerms = [];
        filterState?.filterFields.forEach((element) => {
            let term = {
                name: element.name,
                fields: []
            };
            if (element.filterType === filterType.CHECKBOX) {
                element.fields.forEach((field) => {
                    if (field.checked) {
                        let t = term.fields;
                        t.push(field.name);
                        term.fields = t;
                    }
                });
            } else if (element.filterType === filterType.RADIOBUTTON && element.radioValue !== -1) {
                let t = [];
                t.push(element.fields[element.radioValue].name);
                term.fields = t;
            } else {
                element.fields.forEach((field) => {
                    if (field?.value !== undefined && field?.value !== '') {
                        let t = term.fields;
                        let d = {};
                        d[field.name] = field.value;
                        t.push(d);
                        term.fields = t;
                    }
                });
            }
            if (term.fields.length > 0) filterTerms.push(term);
        });
        setFilterTerms(filterTerms);
    };

    const handleFilter = () => {
        let objTags = [];
        let objCreated = [];
        let objStorage = [];
        let objType = '';
        let objUsage = null;
        let data = state?.domainList;
        switch (filterComponent) {
            case 'stations':
                if (filterTerms?.find((o) => o.name === 'tags')) {
                    objTags = filterTerms?.find((o) => o.name === 'tags')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) =>
                        objTags?.length > 0 ? item.tags.some((tag) => objTags?.includes(tag.name)) : !item.tags.some((tag) => objTags?.includes(tag.name))
                    );
                }
                if (filterTerms?.find((o) => o.name === 'created')) {
                    objCreated = filterTerms?.find((o) => o.name === 'created')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) =>
                        objCreated?.length > 0 ? objCreated?.includes(item.station.created_by_user) : !objCreated?.includes(item.station.created_by_user)
                    );
                }
                if (filterTerms?.find((o) => o.name === 'storage')) {
                    objStorage = filterTerms?.find((o) => o.name === 'storage')?.fields?.map((element) => element?.toLowerCase());
                    data = data.filter((item) =>
                        objStorage?.length > 0 ? objStorage?.includes(item.station.storage_type) : !objStorage?.includes(item.station.storage_type)
                    );
                }
                if (searchInput !== '' && searchInput?.length >= 2) {
                    data = data.filter((station) => station.station.name.includes(searchInput));
                }
                dispatch({ type: 'SET_FILTERED_LIST', payload: data });
                return;
            case 'schemaverse':
                if (filterTerms?.find((o) => o.name === 'tags')) {
                    objTags = filterTerms?.find((o) => o.name === 'tags')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) =>
                        objTags?.length > 0 ? item.tags.some((tag) => objTags?.includes(tag.name)) : !item.tags.some((tag) => objTags?.includes(tag.name))
                    );
                }
                if (filterTerms?.find((o) => o.name === 'created')) {
                    objCreated = filterTerms?.find((o) => o.name === 'created')?.fields?.map((element) => element?.toLowerCase());
                    data = data?.filter((item) => (objCreated?.length > 0 ? objCreated?.includes(item.created_by_user) : !objCreated?.includes(item.created_by_user)));
                }
                if (filterTerms?.find((o) => o.name === 'type')) {
                    objType = filterTerms?.find((o) => o.name === 'type')?.fields[0];
                    data = data?.filter((item) => objType !== '' && item.type === objType);
                }
                if (filterTerms?.find((o) => o.name === 'usage')) {
                    objUsage = filterTerms?.find((o) => o.name === 'usage')?.fields[0] === 'used';
                    data = data.filter((item) => item.used === objUsage);
                }
                if (searchInput !== '' && searchInput?.length >= 2) {
                    data = data.filter((schema) => schema.name.includes(searchInput));
                }
                dispatch({ type: 'SET_FILTERED_LIST', payload: data });
                return;
            default:
                return;
        }
    };

    const handleApply = () => {
        if (filterComponent === 'syslogs') {
            const selectedField = filterState?.filterFields[0]?.radioValue;
            if (selectedField !== -1) {
                dispatch({ type: 'SET_LOG_FILTER', payload: filterState?.filterFields[0]?.fields[selectedField]?.name });
                applyFilter(filterState?.filterFields[0]?.fields[selectedField]?.name);
                setFilterTerms([filterState?.filterFields[0]?.fields[selectedField]?.name]);
            }
        } else createFieldsToFilterObj();
        flipOpen();
    };

    const handleClear = () => {
        filterDispatch({ type: 'SET_COUNTER', payload: 0 });
        let clearFilter = filterFields;
        clearFilter.map((filterGroup) => {
            switch (filterGroup.filterType) {
                case filterType.CHECKBOX:
                    filterGroup.fields.map((field) => (field.checked = false));
                case filterType.DATE:
                    filterGroup.fields.map((field) => (field.value = ''));
                case filterType.RADIOBUTTON:
                    filterGroup.radioValue = -1;
            }
        });
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: clearFilter });
        setFilterTerms([]);
        if (filterComponent === 'syslogs') {
            dispatch({ type: 'SET_LOG_FILTER', payload: 'external' });
            applyFilter('external');
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
                <SearchInput
                    placeholder="Search"
                    colorType="navy"
                    backgroundColorType="gray-dark"
                    width="288px"
                    height="34px"
                    borderColorType="none"
                    boxShadowsType="none"
                    borderRadiusType="circle"
                    iconComponent={<img src={searchIcon} alt="searchIcon" />}
                    onChange={handleSearch}
                    value={searchInput}
                />
            )}
            <Popover className="filter-menu" placement="bottomLeft" content={content} trigger="click" onOpenChange={handleOpenChange} open={filterState.isOpen}>
                <Button
                    className="modal-btn"
                    width="110px"
                    height={height}
                    placeholder={
                        <div className="filter-container">
                            <img src={filterImg} width="25" alt="filter" />
                            <label className="filter-title">Filters</label>
                            {filterTerms?.length > 0 && filterState?.counter > 0 && <div className="filter-counter">{filterTerms?.length || filterState?.counter}</div>}
                        </div>
                    }
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="14px"
                    fontWeight="bold"
                    boxShadowStyle="float"
                    onClick={() => {}}
                />
            </Popover>
        </FilterStoreContext.Provider>
    );
};
export const FilterStoreContext = createContext({});
export default Filter;
