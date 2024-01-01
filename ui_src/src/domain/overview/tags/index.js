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
import Tag from 'components/tag';
import { Divider } from 'antd';
import { Context } from 'hooks/store';
import { ReactComponent as NoTagsFoundIcon } from 'assets/images/noTagsFound.svg';

const Tags = () => {
    const [state, dispatch] = useContext(Context);
    return (
        <div className="overview-components-wrapper">
            {state?.monitor_data?.tags_details?.length > 0 ? (
                <div className="overview-tags-container">
                    <div className="overview-components-header">
                        <p>Most used tags</p>
                    </div>
                    <div className="tags-items-container">
                        {state?.monitor_data?.tags_details?.map((tag, index) => {
                            return (
                                <div key={index}>
                                    <span className="tag-item">
                                        <span className="item">
                                            <label className="item-num">{`${index + 1}`}</label>
                                            <Tag tag={{ color: tag.color, name: tag.name }} />
                                        </span>
                                        <label className="attached-component">
                                            {`${tag.stations_count} station${tag.stations_count === 1 ? '' : 's'}, ${tag.schemas_count} schema${
                                                tag.schemas_count === 1 ? '' : 's'
                                            }`}
                                        </label>
                                    </span>
                                    <Divider />
                                </div>
                            );
                        })}
                    </div>
                </div>
            ) : (
                <div className="no-data">
                    <NoTagsFoundIcon alt="no data found" />
                    <p>No tags yet</p>
                    <label>Tags are a great way to identify your different assets as you grow.</label>
                </div>
            )}
        </div>
    );
};

export default Tags;
