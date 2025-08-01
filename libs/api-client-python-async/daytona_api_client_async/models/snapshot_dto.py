# coding: utf-8

"""
    Daytona

    Daytona AI platform API Docs

    The version of the OpenAPI document: 1.0
    Contact: support@daytona.com
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from __future__ import annotations
import pprint
import re  # noqa: F401
import json

from datetime import datetime
from pydantic import BaseModel, ConfigDict, Field, StrictBool, StrictFloat, StrictInt, StrictStr
from typing import Any, ClassVar, Dict, List, Optional, Union
from daytona_api_client_async.models.build_info import BuildInfo
from daytona_api_client_async.models.snapshot_state import SnapshotState
from typing import Optional, Set
from typing_extensions import Self

class SnapshotDto(BaseModel):
    """
    SnapshotDto
    """ # noqa: E501
    id: StrictStr
    organization_id: Optional[StrictStr] = Field(default=None, alias="organizationId")
    general: StrictBool
    name: StrictStr
    image_name: Optional[StrictStr] = Field(default=None, alias="imageName")
    state: SnapshotState
    size: Optional[Union[StrictFloat, StrictInt]]
    entrypoint: Optional[List[StrictStr]]
    cpu: Union[StrictFloat, StrictInt]
    gpu: Union[StrictFloat, StrictInt]
    mem: Union[StrictFloat, StrictInt]
    disk: Union[StrictFloat, StrictInt]
    error_reason: Optional[StrictStr] = Field(alias="errorReason")
    created_at: datetime = Field(alias="createdAt")
    updated_at: datetime = Field(alias="updatedAt")
    last_used_at: Optional[datetime] = Field(alias="lastUsedAt")
    build_info: Optional[BuildInfo] = Field(default=None, description="Build information for the snapshot", alias="buildInfo")
    additional_properties: Dict[str, Any] = {}
    __properties: ClassVar[List[str]] = ["id", "organizationId", "general", "name", "imageName", "state", "size", "entrypoint", "cpu", "gpu", "mem", "disk", "errorReason", "createdAt", "updatedAt", "lastUsedAt", "buildInfo"]

    model_config = ConfigDict(
        populate_by_name=True,
        validate_assignment=True,
        protected_namespaces=(),
    )


    def to_str(self) -> str:
        """Returns the string representation of the model using alias"""
        return pprint.pformat(self.model_dump(by_alias=True))

    def to_json(self) -> str:
        """Returns the JSON representation of the model using alias"""
        # TODO: pydantic v2: use .model_dump_json(by_alias=True, exclude_unset=True) instead
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> Optional[Self]:
        """Create an instance of SnapshotDto from a JSON string"""
        return cls.from_dict(json.loads(json_str))

    def to_dict(self) -> Dict[str, Any]:
        """Return the dictionary representation of the model using alias.

        This has the following differences from calling pydantic's
        `self.model_dump(by_alias=True)`:

        * `None` is only added to the output dict for nullable fields that
          were set at model initialization. Other fields with value `None`
          are ignored.
        * Fields in `self.additional_properties` are added to the output dict.
        """
        excluded_fields: Set[str] = set([
            "additional_properties",
        ])

        _dict = self.model_dump(
            by_alias=True,
            exclude=excluded_fields,
            exclude_none=True,
        )
        # override the default output from pydantic by calling `to_dict()` of build_info
        if self.build_info:
            _dict['buildInfo'] = self.build_info.to_dict()
        # puts key-value pairs in additional_properties in the top level
        if self.additional_properties is not None:
            for _key, _value in self.additional_properties.items():
                _dict[_key] = _value

        # set to None if size (nullable) is None
        # and model_fields_set contains the field
        if self.size is None and "size" in self.model_fields_set:
            _dict['size'] = None

        # set to None if entrypoint (nullable) is None
        # and model_fields_set contains the field
        if self.entrypoint is None and "entrypoint" in self.model_fields_set:
            _dict['entrypoint'] = None

        # set to None if error_reason (nullable) is None
        # and model_fields_set contains the field
        if self.error_reason is None and "error_reason" in self.model_fields_set:
            _dict['errorReason'] = None

        # set to None if last_used_at (nullable) is None
        # and model_fields_set contains the field
        if self.last_used_at is None and "last_used_at" in self.model_fields_set:
            _dict['lastUsedAt'] = None

        return _dict

    @classmethod
    def from_dict(cls, obj: Optional[Dict[str, Any]]) -> Optional[Self]:
        """Create an instance of SnapshotDto from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return cls.model_validate(obj)

        _obj = cls.model_validate({
            "id": obj.get("id"),
            "organizationId": obj.get("organizationId"),
            "general": obj.get("general"),
            "name": obj.get("name"),
            "imageName": obj.get("imageName"),
            "state": obj.get("state"),
            "size": obj.get("size"),
            "entrypoint": obj.get("entrypoint"),
            "cpu": obj.get("cpu"),
            "gpu": obj.get("gpu"),
            "mem": obj.get("mem"),
            "disk": obj.get("disk"),
            "errorReason": obj.get("errorReason"),
            "createdAt": obj.get("createdAt"),
            "updatedAt": obj.get("updatedAt"),
            "lastUsedAt": obj.get("lastUsedAt"),
            "buildInfo": BuildInfo.from_dict(obj["buildInfo"]) if obj.get("buildInfo") is not None else None
        })
        # store additional fields in additional_properties
        for _key in obj.keys():
            if _key not in cls.__properties:
                _obj.additional_properties[_key] = obj.get(_key)

        return _obj


