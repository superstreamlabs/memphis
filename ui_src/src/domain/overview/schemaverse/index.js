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

import React, { useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { Divider } from 'antd';
import pathDomains from '../../../router';
import SchemaChart from './schemaChart';
import { Context } from '../../../hooks/store';
import noSchemasFound from '../../../assets/images/noSchemasFound.svg';

const Schemaverse = () => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    return (
        <div className="overview-components-wrapper">
            {state?.monitor_data?.schemas_details?.total_schemas > 0 ? (
                <div className="overview-schema-container">
                    <div className="overview-components-header schemaverse-header">
                        <p> Schemaverse </p>
                        <label className="link-to-page" onClick={() => history.push(`${pathDomains.schemaverse}/list`)}>
                            Go to schemaverse
                        </label>
                    </div>
                    <div className="total-data sum">
                        <span>
                            <p className="total-measure">Total schemas</p>
                            <p className="total-value">{state?.monitor_data?.schemas_details?.total_schemas}</p>
                        </span>
                        <Divider type="vertical" />
                        <span>
                            <p className="total-measure">Enforced schemas</p>
                            <p className="total-value">{state?.monitor_data?.schemas_details?.enforced_schemas}</p>
                        </span>
                    </div>
                    <div className="total-data info">
                        <SchemaChart
                            schemas={[
                                { name: 'Protobuf', usage: state?.monitor_data?.schemas_details?.protobuf || 0 },
                                { name: 'Json', usage: state?.monitor_data?.schemas_details?.json_schema || 0 },
                                { name: 'GraphQL', usage: state?.monitor_data?.schemas_details?.Graphql || 0 },
                                { name: 'Avro', usage: state?.monitor_data?.schemas_details?.avro || 0 }
                            ]}
                        />
                    </div>
                </div>
            ) : (
                <div className="no-data">
                    <img src={noSchemasFound} alt="no data found" />
                    <p>No schemas yet</p>
                    <label>Schemas are made to force producers to produce messages in a specific structure and format.</label>
                    <label className="link" onClick={() => history.push(`${pathDomains.schemaverse}/create`)}>
                        + Create a schema
                    </label>
                </div>
            )}
        </div>
    );
};

export default Schemaverse;
