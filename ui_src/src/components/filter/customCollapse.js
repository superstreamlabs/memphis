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

import React, { useState, useContext, useEffect } from 'react';
import { FilterStoreContext } from './';

import { Collapse } from 'antd';
import { Checkbox } from 'antd';
import { Divider } from 'antd';
import Tag from '../tag';

import { ReactComponent as CollapseArrowIcon } from '../../assets/images/collapseArrow.svg';
import Button from '../button';
import DatePicker from '../datePicker';
import RadioButton from '../radioButton';
import { filterType, labelType } from '../../const/globalConst';

const { Panel } = Collapse;

const CustomCollapse = ({ cancel, apply, clear }) => {
    const [filterState, filterDispatch] = useContext(FilterStoreContext);
    const [activeKey, setActiveKey] = useState(['0', '1', '2']);

    useEffect(() => {
        if (activeKey?.length > 3) {
            const shortActiveKey = activeKey.splice(0, activeKey?.length - 3);
            setActiveKey(shortActiveKey);
        }
    }, [activeKey]);

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
                        <div className="date-container" key={filterField?.name}>
                            <label>{filterField?.label}</label>
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
                            return { id: id, value: id, label: item?.name };
                        })}
                        radioStyle="radiobtn-capitalize"
                        radioValue={filterGroup.radioValue}
                        onChange={(e) => updateChoice(filterGroupIndex, e.target.value)}
                    />
                );
        }
    };

    const drawCheckBox = (filterGroup, filterGroupIndex) => {
        switch (filterGroup.labelType) {
            case labelType.BADGE:
                return filterGroup?.fields?.map((filterField, filterFieldIndex = 0) => {
                    if (filterFieldIndex < 3)
                        return (
                            <div className="label-container" key={filterField?.name}>
                                <Checkbox
                                    checked={filterField?.checked || false}
                                    onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)}
                                    name={filterGroup?.name}
                                />
                                <Tag tag={{ color: filterField.color, name: filterField?.name }}></Tag>
                            </div>
                        );
                    else {
                        return filterGroup.showMore ? (
                            <div>
                                <div className="label-container" key={filterField.name}>
                                    <Checkbox
                                        checked={filterField?.checked || false}
                                        onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)}
                                        name={filterGroup?.name}
                                    />
                                    <Tag tag={{ color: filterField.color, name: filterField?.name }}></Tag>
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
                return filterGroup?.fields?.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="circle-container" key={filterField?.name}>
                            <Checkbox
                                checked={filterField?.checked || false}
                                onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)}
                                name={filterGroup?.name}
                            />
                            <p className="circle-letter" style={{ backgroundColor: filterField.color }}>
                                {filterField?.name[0]?.toUpperCase()}
                            </p>
                            <label>{filterField?.name}</label>
                        </div>
                    );
                });
            default:
                return filterGroup.fields.map((filterField, filterFieldIndex = 0) => {
                    return (
                        <div className="default-checkbox" key={filterField?.name}>
                            <Checkbox checked={filterField.checked} onChange={() => updateChoice(filterGroupIndex, filterFieldIndex)} name={filterGroup?.name} />
                            <label>{filterField?.name}</label>
                        </div>
                    );
                });
        }
    };

    return (
        <Collapse ghost defaultActiveKey={['0', '1', '2']} onChange={onChange} className="custom-collapse-filter">
            <div className="collapse-header">
                <div className="header-name-counter">
                    <label>Filter</label>
                    {filterState?.counter > 0 && <div className="filter-counter">{filterState?.counter}</div>}
                </div>
                <label className="clear" onClick={clear}>
                    Clear All
                </label>
            </div>
            {filterState?.filterFields?.map((filterGroup, filterGroupIndex = 0) => (
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
                                    <CollapseArrowIcon
                                        className={activeKey?.includes(filterGroupIndex.toString()) ? 'collapse-arrow open' : 'collapse-arrow'}
                                        alt="collapse-arrow"
                                    />
                                </div>
                            </div>
                        )
                    }
                    key={`${filterGroupIndex}`}
                    showArrow={false}
                >
                    <div className="tag-container">{drawComponent(filterGroup, filterGroupIndex)}</div>
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
