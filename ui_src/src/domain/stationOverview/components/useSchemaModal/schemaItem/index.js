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

import { CloseRounded, FiberManualRecord } from '@material-ui/icons';
import React from 'react';
import { ReactComponent as SchemaItemIcon } from '../../../../../assets/images/schemaItemIcon.svg';
import { ReactComponent as StationIcon } from '../../../../../assets/images/stationsIconActive.svg';
import { parsingDate } from '../../../../../services/valueConvertor';

const SchemaItem = ({ schema, handleSelectedItem, selected, type, handleStopUseSchema }) => {
    return (
        <div
            key={schema?.id}
            className={selected === schema?.name ? 'schema-item-container sch-item-selected' : 'schema-item-container'}
            onClick={() => handleSelectedItem(schema?.name)}
        >
            <div className="content">
                <div className="name-wrapper">
                    {type === 'dls' ? <StationIcon /> : <SchemaItemIcon />}

                    <p className="name">{schema?.name}</p>
                </div>
                <div className="details">
                    <p className="type">{schema?.type}</p>
                    <FiberManualRecord />
                    <p className="date">{parsingDate(schema?.created_at)}</p>
                </div>
            </div>
        </div>
    );
};

export default SchemaItem;
