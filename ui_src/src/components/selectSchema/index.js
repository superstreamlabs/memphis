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

import { ReactComponent as SchemaSelectIcon } from '../../assets/images/schemaIconSelect.svg';
import { ReactComponent as PlaceholderSchemaIcon } from '../../assets/images/placeholderSchema.svg';
import { parsingDate } from '../../services/valueConvertor';
import { Context } from '../../hooks/store';
import Button from '../button';
import pathDomains from '../../router';

const { Option } = Select;

const SelectSchema = ({ options, onChange, value, placeholder }) => {
    const history = useHistory();

    const handleChange = (e) => {
        onChange(e);
    };

    const createNew = () => {
        history.push(`${pathDomains.schemaverse}/create`);
    };

    return (
        <div className="select-schema-container">
            <Select
                className="select"
                placeholder={placeholder}
                value={value}
                bordered={false}
                suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                onChange={handleChange}
                placement="bottomRight"
                popupClassName="select-schema-options"
                notFoundContent={
                    <div className="no-schema-to-display">
                        <PlaceholderSchemaIcon alt="PlaceholderSchemaIcon" width={50} height={50} />
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
                }
            >
                {options?.map((schema) => {
                    return (
                        <Option key={schema?.id} value={schema?.name}>
                            <div className="schema-details">
                                <SchemaSelectIcon alt="SchemaSelectIcon" height={20} />
                                <p className="schema-name">{schema?.name}</p>
                            </div>
                            <p className="created-by">
                                {schema?.type} &#8226; {parsingDate(schema?.created_at)}
                            </p>
                        </Option>
                    );
                })}
            </Select>
        </div>
    );
};

export default SelectSchema;
