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

import { ArrowDropDownRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import React, { useContext } from 'react';
import { Select } from 'antd';

import SchemaIconSelect from '../../assets/images/schemaIconSelect.svg';
import plus from '../../assets/images/plus.svg';
import placeholderSchema from '../../assets/images/placeholderSchema.svg';
import usersIconActive from '../../assets/images/usersIconActive.svg';
import { parsingDate } from '../../services/valueConvertor';
import { Context } from '../../hooks/store';
import Button from '../button';
import pathDomains from '../../router';

const { Option } = Select;

const CustomSelect = ({ options, onChange, value, placeholder, type = 'schema', handleCreateNew }) => {
    const history = useHistory();

    const handleChange = (e) => {
        onChange(e);
    };

    const createNewSchema = () => {
        history.push(`${pathDomains.schemaverse}/create`);
    };

    const displayValue = value === '' || value === null ? null : value;

    return (
        <div className="select-schema-container">
            <Select
                className="select"
                placeholder={placeholder}
                value={displayValue}
                bordered={false}
                suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                onChange={handleChange}
                placement="bottomRight"
                popupClassName="select-schema-options"
                notFoundContent={
                    <div className="no-schema-to-display">
                        <div className="top">
                            <p className="no-result-found">No Result Found</p>
                        </div>
                        <div className="divider" />
                        <div className="bottom">
                            <Button
                                placeholder={
                                    <div className="create-btn">
                                        <img src={plus} alt="verifiedIcon" />
                                        <p>{type === 'schema' ? 'Create schema' : type === 'user' ? 'Create user' : ''}</p>
                                    </div>
                                }
                                className="modal-btn"
                                width="83px"
                                height="32px"
                                colorType="purple"
                                radiusType="circle"
                                backgroundColorType={'none'}
                                fontSize="12px"
                                fontWeight="600"
                                onClick={() => {
                                    return type === 'schema' ? createNewSchema() : type === 'user' ? handleCreateNew() : null;
                                }}
                            />
                        </div>
                    </div>
                }
            >
                {options?.map((schema) => {
                    return (
                        <Option key={schema?.id} value={schema?.name}>
                            {type === 'schema' ? (
                                <>
                                    <div className="schema-details">
                                        <img src={SchemaIconSelect} alt="SchemaIconSelect" height="20px" />
                                        <p className="schema-name">{schema?.name}</p>
                                    </div>
                                    <p className="created-by">
                                        {schema?.type} &#8226; {parsingDate(schema?.created_at)}
                                    </p>
                                </>
                            ) : type === 'user' ? (
                                <>
                                    <div className="schema-details">
                                        <img src={usersIconActive} alt="usersIcon" height={20} width={20} />
                                        <p className="schema-name">{schema?.name}</p>
                                    </div>
                                    <p className="created-by">{parsingDate(schema?.created_at)}</p>
                                </>
                            ) : null}
                        </Option>
                    );
                })}
            </Select>
        </div>
    );
};

export default CustomSelect;
