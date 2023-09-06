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

    const createNew = () => {
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
                    type === 'schema' ? (
                        <div className="no-schema-to-display">
                            <img src={placeholderSchema} width="50" height="50" alt="placeholderSchema" />
                            <p className="title">No schemas yet</p>
                            <p className="sub-title">Get started by creating your first schema</p>
                            <Button
                                className="modal-btn"
                                width="120px"
                                height="34px"
                                placeholder="Create schema"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                aria-controls="usecse-menu"
                                aria-haspopup="true"
                                onClick={() => createNew()}
                            />
                        </div>
                    ) : type === 'user' ? (
                        <div className="no-schema-to-display">
                            <div className="placeholder-background">
                                <img src={usersIconActive} width={40} height={40} alt="placeholderSchema" />
                            </div>
                            <p className="title">No users yet</p>
                            <p className="sub-title">Get started by creating your first user</p>
                            <Button
                                className="modal-btn"
                                width="120px"
                                height="34px"
                                placeholder="Create user"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                aria-controls="usecse-menu"
                                aria-haspopup="true"
                                onClick={handleCreateNew}
                            />
                        </div>
                    ) : null
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
