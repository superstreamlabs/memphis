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

import React, { useState, useContext, useEffect } from 'react';
import { FilterStoreContext } from './';

import { Collapse } from 'antd';
import { Checkbox } from 'antd';
import { Divider } from 'antd';
import Tag from '../../components/tag';

import CollapseArrow from '../../assets/images/collapseArrow.svg';
import Button from '../button';
import DatePicker from '../datePicker';
import RadioButton from '../radioButton';
import { filterType, labelType } from '../../const/globalConst';

const { Panel } = Collapse;

const CustomCollapse = ({ cancel, apply, clear }) => {
    const [filterState, filterDispatch] = useContext(FilterStoreContext);
    const [activeKey, setActiveKey] = useState(['0']);

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                apply();
            }
        };
        document.addEventListener('keydown', keyDownHandler);
        return () => {
            document.removeEventListener('keydown', keyDownHandler);
        };
    }, []);

    const onChange = (key) => {
        setActiveKey(key);
    };

    const updateChoice = (filterGroup, filterField, value) => {
        let updatedCounter;
        let filter;
        updatedCounter = filterState.counter;
        filter = filterState.filterFields;
        switch (filterState.filterFields[filterGroup].filterType) {
            case filterType.CHECKBOX:
                if (filter[filterGroup].fields[filterField].checked) updatedCounter--;
                else updatedCounter++;
                filter[filterGroup].fields[filterField].checked = !filter[filterGroup].fields[filterField].checked;
                break;
            case filterType.RADIOBUTTON:
                if (filter[filterGroup].radioValue === -1) updatedCounter++;
                filter[filterGroup].radioValue = filterField;
                break;
            case filterType.DATE:
                if (filter[filterGroup].fields[filterField].value === '' && value !== '') updatedCounter++;
                else if (filter[filterGroup].fields[filterField].value !== '' && value === '') updatedCounter--;
                filter[filterGroup].fields[filterField].value = value;
                break;
        }
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filter });
        filterDispatch({ type: 'SET_COUNTER', payload: updatedCounter });
    };

    const showMoreLess = (index, showMoreFalg) => {
        let filter = filterState.filterFields;
        filter[index].showMore = showMoreFalg;
        filterDispatch({ type: 'SET_FILTER_FIELDS', payload: filter });
    };

    const drawComponent = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.filterType) {
            case filterType.CHECKBOX:
                return drawCheckBox(filterGroup, filterGroupIndex);
            case filterType.DATE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="date-container" key={filterField.name}>
                            <label>{filterField.label}</label>
                            <DatePicker
                                type="text"
                                radiusType="semi-round"
                                colorType="gray"
                                backgroundColorType="none"
                                borderColorType="red"
                                width="240px"
                                minWidth="200px"
                                onChange={(e) => updateChoice(filterGroupIndex, filterFieldIndex, e)}
                            />
                        </div>
                    );
                });
            case filterType.RADIOBUTTON:
                return (
                    <RadioButton
                        vertical={true}
                        height="25px"
                        fontFamily="InterSemiBold"
                        options={filterGroup.fields.map((item, id) => {
                            return { id: id, value: id, label: item.name };
                        })}
                        radioValue={filterGroup.radioValue}
                        onChange={(e) => updateChoice(filterGroupIndex, e.target.value)}
                    />
                );
        }
    };

    const drawCheckBox = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.labelType) {
            case labelType.BADGE:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    if (filterFieldIndex < 3)
                        return (
                            <div className="label-container" key={filterField.name}>
                                <Checkbox checked={filterField.checked} onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                                <Tag tag={{ color: filterField.color, name: filterField.name }}></Tag>
                            </div>
                        );
                    else {
                        return filterGroup.showMore ? (
                            <div>
                                <div className="label-container" key={filterField.name}>
                                    <Checkbox checked={filterField.checked} onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                                    <Tag tag={{ color: filterField.color, name: filterField.name }}></Tag>
                                </div>
                                {filterFieldIndex === filterGroup.fields.length - 1 && (
                                    <p className="show-more" onClick={() => showMoreLess(filterGroupIndex, false)}>
                                        Show Less...
                                    </p>
                                )}
                            </div>
                        ) : (
                            filterFieldIndex === 3 && (
                                <p className="show-more" onClick={() => showMoreLess(filterGroupIndex, true)}>
                                    Show All...
                                </p>
                            )
                        );
                    }
                });
            case labelType.CIRCLEDLETTER:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="circle-container" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                            <p className="circle-letter" style={{ backgroundColor: filterField.color }}>
                                {filterField.name[0]?.toUpperCase()}
                            </p>
                            <label>{filterField.name}</label>
                        </div>
                    );
                });
            default:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="default-checkbox" key={filterField.name}>
                            <Checkbox checked={filterField.checked} onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)} name={filterGroup.name} />
                            <label>{filterField.name}</label>
                        </div>
                    );
                });
        }
    };

    return (
        <Collapse ghost defaultActiveKey={['0']} onChange={onChange} className="custom-collapse-filter">
            <div className="collapse-header">
                <div className="header-name-counter">
                    <label>Filter</label>
                    {filterState?.counter > 0 && <div className="filter-counter">{filterState?.counter}</div>}
                </div>
                <label className="clear" onClick={clear}>
                    Clear All
                </label>
            </div>
            {filterState?.filterFields.map((filterGroup, filterGroupIndex = 0) => (
                <Panel
                    header={
                        filterGroup?.fields?.length > 0 && (
                            <div>
                                {filterGroupIndex !== 0 && (
                                    <div className="divider-container">
                                        <Divider />
                                    </div>
                                )}
                                <div className="filter-header">
                                    <label className="title">{filterGroup.value}</label>
                                    <img
                                        className={activeKey?.includes(filterGroupIndex.toString()) ? 'collapse-arrow open' : 'collapse-arrow'}
                                        src={CollapseArrow}
                                        alt="collapse-arrow"
                                    />
                                </div>
                            </div>
                        )
                    }
                    key={`${filterGroupIndex}`}
                    showArrow={false}
                >
                    <div className={filterGroup.labelType === labelType.BADGE ? 'tag-container tag-container-gray' : 'tag-container'}>
                        {drawComponent(filterGroup, filterGroupIndex)}
                    </div>
                </Panel>
            ))}

            <div className="collapse-footer">
                <Button
                    width="110px"
                    height="30px"
                    placeholder="Cancel"
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType={'white'}
                    border={'gray'}
                    fontSize="12px"
                    fontWeight="bold"
                    onClick={cancel}
                />
                <Button
                    width="110px"
                    height="30px"
                    placeholder="Apply"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType={'purple'}
                    fontSize="12px"
                    fontWeight="bold"
                    onClick={apply}
                />
            </div>
        </Collapse>
    );
};
export default CustomCollapse;
